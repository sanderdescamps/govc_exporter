package scraper

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper/logger"
	sensormetrics "github.com/sanderdescamps/govc_exporter/internal/scraper/sensor_metrics"
	"github.com/vmware/govmomi/vim25/types"
)

const HOST_PERF_SENSOR_NAME = "HostPerfSensor"

func DefaultHostPerfMetrics() []string {
	return []string{"cpu.usagemhz.average",
		"cpu.demand.average",
		"cpu.latency.average",
		"cpu.entitlement.latest",
		"cpu.ready.summation",
		"cpu.readiness.average",
		"cpu.costop.summation",
		"cpu.maxlimited.summation",
		"mem.entitlement.average",
		"mem.active.average",
		"mem.shared.average",
		"mem.vmmemctl.average",
		"mem.swapped.average",
		"mem.consumed.average",
		"net.bytesRx.average",
		"net.bytesTx.average",
		"net.errorsRx.summation",
		"net.errorsTx.summation",
		"net.droppedRx.summation",
		"net.droppedTx.summation",
		"datastore.read.average",
		"datastore.write.average",
		// "datastore.numberReadAveraged.average",
		// "datastore.numberWriteAveraged.average",
		// "datastore.totalReadLatency.average",
		// "datastore.totalWriteLatency.average",
	}
}

type HostPerfSensor struct {
	BasePerfSensor
	logger.SensorLogger
	metricsCollector *sensormetrics.SensorMetricsCollector
	statusMonitor    *sensormetrics.StatusMonitor
	started          *helper.StartedCheck
	sensorLock       sync.Mutex
	manualRefresh    chan struct{}
	stopChan         chan struct{}
	config           config.PerfSensorConfig
}

func NewHostPerfSensor(scraper *VCenterScraper, config config.PerfSensorConfig, l *slog.Logger) *HostPerfSensor {
	var mc *sensormetrics.SensorMetricsCollector = sensormetrics.NewLastSensorMetricsCollector()
	var sm *sensormetrics.StatusMonitor = sensormetrics.NewStatusMonitor()

	metrics := []string{}
	if config.DefaultMetrics {
		metrics = append(metrics, DefaultHostPerfMetrics()...)
	}
	metrics = append(metrics, config.ExtraMetrics...)
	metrics = helper.Dedup(metrics)

	var sensor HostPerfSensor = HostPerfSensor{
		BasePerfSensor:   *NewBasePerfSensor(config, metrics, mc, sm),
		config:           config,
		started:          helper.NewStartedCheck(),
		stopChan:         make(chan struct{}),
		manualRefresh:    make(chan struct{}),
		SensorLogger:     logger.NewSLogLogger(l, logger.WithKind(HOST_PERF_SENSOR_NAME)),
		metricsCollector: mc,
		statusMonitor:    sm,
	}
	return &sensor
}

func (s *HostPerfSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	(scraper.Host).(*HostSensor).WaitTillStartup()

	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	oHostRefs := scraper.DB.GetAllHostRefs(ctx)
	if len(oHostRefs) < 1 {
		s.SensorLogger.Info("No hosts found, no host perf metrics available", "refs", oHostRefs)
		return nil
	}

	var hostRefs []types.ManagedObjectReference
	for _, ref := range oHostRefs {
		hostRefs = append(hostRefs, ref.ToVMwareRef())
	}

	metricSeries, err := s.BasePerfSensor.QueryEntiryMetrics(ctx, scraper, hostRefs)
	if err != nil {
		return err
	}
	for _, metricSerie := range metricSeries {
		entityRef := objects.NewManagedObjectReferenceFromVMwareRef(metricSerie.Entity)
		metrics := EntityMetricToMetric(metricSerie)
		err := scraper.MetricsDB.AddHostMetrics(ctx, entityRef, s.config.MaxAge, metrics...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *HostPerfSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
	if !s.started.IsStarted() {
		err := s.refresh(ctx, scraper)
		if err != nil {
			s.statusMonitor.Fail()
			return err
		}
		s.statusMonitor.Success()
		s.started.Started()
	} else {
		return ErrSensorAlreadyStarted
	}
	return nil
}

func (s *HostPerfSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
	ticker := time.NewTicker(s.config.RefreshInterval)
	go func() {
		time.Sleep(time.Duration(rand.Intn(20000)) * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				go func() {
					err := s.refresh(ctx, scraper)
					if err == nil {
						s.SensorLogger.Info("refresh successful")
						s.statusMonitor.Success()
					} else {
						s.SensorLogger.Error("refresh failed", "err", err)
						s.statusMonitor.Fail()
					}
				}()
			case <-s.manualRefresh:
				go func() {
					s.SensorLogger.Info("trigger manual refresh")
					err := s.refresh(ctx, scraper)
					if err == nil {
						s.SensorLogger.Info("manual refresh successful")
						s.statusMonitor.Success()
					} else {
						s.SensorLogger.Error("manual refresh failed", "err", err)
						s.statusMonitor.Fail()
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

func (s *HostPerfSensor) StopRefresher(ctx context.Context) {
	close(s.stopChan)
}

func (s *HostPerfSensor) TriggerManualRefresh(ctx context.Context) {
	s.manualRefresh <- struct{}{}
}

func (s *HostPerfSensor) Kind() string {
	return "HostPerfSensor"
}

func (s *HostPerfSensor) WaitTillStartup() {
	s.started.Wait()
}

func (s *HostPerfSensor) Match(name string) bool {
	return helper.NewMatcher(
		"perf-host",
		"perfhost",
		"perfesx",
		"perf-esx",
		"host-perf",
		"hostperf",
		"esxperf",
		"esx-perf").Match(name)
}

func (s *HostPerfSensor) Enabled() bool {
	return true
}

func (s *HostPerfSensor) GetLatestMetrics() []sensormetrics.SensorMetric {
	return append(
		s.metricsCollector.ComposeMetrics(s.Kind()),
		sensormetrics.SensorMetric{
			Sensor:     s.Kind(),
			MetricName: "failed",
			Value:      s.statusMonitor.StatusFailedFloat64(),
			Unit:       "boolean",
		}, sensormetrics.SensorMetric{
			Sensor:     s.Kind(),
			MetricName: "fail_rate",
			Value:      s.statusMonitor.FailRate(),
			Unit:       "boolean",
		}, sensormetrics.SensorMetric{
			Sensor:     s.Kind(),
			MetricName: "enabled",
			Value:      1.0,
			Unit:       "boolean",
		},
	)
}
