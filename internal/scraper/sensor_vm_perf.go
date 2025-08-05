package scraper

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper/logger"
	metricshelper "github.com/sanderdescamps/govc_exporter/internal/scraper/metrics_helper"
	"github.com/vmware/govmomi/vim25/types"
)

const VM_PERF_SENSOR_NAME = "VMPerfSensor"

func DefaultVMPerfMetrics() []string {
	return []string{
		"cpu.usagemhz.average",
		"cpu.usagemhz.average",
		"cpu.capacity.provisioned.average",
		"cpu.readiness.average",
		"cpu.costop.summation",
		"cpu.maxlimited.summation",
		"cpu.ready.summation",
		"mem.active.average",
		"mem.granted.average",
		"mem.consumed.average",
		"disk.throughput.contention.average",
		"disk.throughput.usage.average",
	}
}

type VMPerfSensor struct {
	BasePerfSensor
	logger.SensorLogger
	metricshelper.MetricHelperDefault
	started       *helper.StartedCheck
	sensorLock    sync.Mutex
	manualRefresh chan struct{}
	stopChan      chan struct{}
	config        PerfSensorConfig
}

func NewVMPerfSensor(scraper *VCenterScraper, config PerfSensorConfig, l *slog.Logger) *VMPerfSensor {
	metrics := []string{}
	if config.DefaultMetrics {
		metrics = append(metrics, DefaultVMPerfMetrics()...)
	}
	metrics = append(metrics, config.ExtraMetrics...)
	metrics = helper.Dedup(metrics)

	var sensor VMPerfSensor = VMPerfSensor{
		BasePerfSensor:      *NewBasePerfSensor(config, metrics),
		started:             helper.NewStartedCheck(),
		stopChan:            make(chan struct{}),
		manualRefresh:       make(chan struct{}),
		config:              config,
		SensorLogger:        logger.NewSLogLogger(l, logger.WithKind(VM_PERF_SENSOR_NAME)),
		MetricHelperDefault: *metricshelper.NewMetricHelperDefault(VM_PERF_SENSOR_NAME),
	}
	return &sensor
}

func (s *VMPerfSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	(scraper.VM).(*VirtualMachineSensor).WaitTillStartup()

	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	oVMRefs := scraper.DB.GetAllVMRefs(ctx)
	if len(oVMRefs) < 1 {
		s.SensorLogger.Info("No VMs found, no vm perf metrics available")
		return nil
	}

	var vmRefs []types.ManagedObjectReference
	for _, ref := range oVMRefs {
		vmRefs = append(vmRefs, ref.ToVMwareRef())
	}

	metricSeries, refreshStats, err := s.BasePerfSensor.QueryEntiryMetrics(ctx, scraper, vmRefs)
	s.MetricHelperDefault.LoadStats(refreshStats)
	if err != nil {
		return err
	}
	for _, metricSerie := range metricSeries {
		entityRef := objects.NewManagedObjectReferenceFromVMwareRef(metricSerie.Entity)
		metrics := EntityMetricToMetric(metricSerie)
		err := scraper.MetricsDB.AddVmMetrics(ctx, entityRef, metrics...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *VMPerfSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
	if !s.started.IsStarted() {
		err := s.refresh(ctx, scraper)
		if err != nil {
			return err
		}
		s.started.Started()
	} else {
		return ErrSensorAlreadyStarted
	}
	return nil
}

func (s *VMPerfSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
	ticker := time.NewTicker(s.config.RefreshInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				go func() {
					err := s.refresh(ctx, scraper)
					if err == nil {
						s.SensorLogger.Info("refresh successful")
					} else {
						s.SensorLogger.Error("refresh failed", "err", err)
					}
				}()
			case <-s.manualRefresh:
				go func() {
					s.SensorLogger.Info("trigger manual refresh")
					err := s.refresh(ctx, scraper)
					if err == nil {
						s.SensorLogger.Info("manual refresh successful")
					} else {
						s.SensorLogger.Error("manual refresh failed", "err", err)
					}
				}()
			case <-s.stopChan:
				s.started.Stopped()
				ticker.Stop()
			case <-ctx.Done():
				s.started.Stopped()
				ticker.Stop()
			}
		}
	}()
	return nil
}

func (s *VMPerfSensor) StopRefresher(ctx context.Context) {
	close(s.stopChan)
}

func (s *VMPerfSensor) TriggerManualRefresh(ctx context.Context) {
	s.manualRefresh <- struct{}{}
}

func (s *VMPerfSensor) Kind() string {
	return "VMPerfSensor"
}

func (s *VMPerfSensor) WaitTillStartup() {
	s.started.Wait()
}

func (s *VMPerfSensor) Match(name string) bool {
	return helper.NewMatcher("perf-vm", "perfvm", "vm-perf", "vmperf").Match(name)
}

func (s *VMPerfSensor) Enabled() bool {
	return true
}
