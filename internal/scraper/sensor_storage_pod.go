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
	"github.com/vmware/govmomi/vim25/mo"
)

const STORAGE_POD_SENSOR_NAME = "StoragePodSensor"

type StoragePodSensor struct {
	BaseSensor
	logger.SensorLogger
	metricsCollector *sensormetrics.SensorMetricsCollector
	statusMonitor    *sensormetrics.StatusMonitor
	started          *helper.StartedCheck
	sensorLock       sync.Mutex
	manualRefresh    chan struct{}
	stopChan         chan struct{}
	config           config.SensorConfig
}

func NewStoragePodSensor(scraper *VCenterScraper, config config.SensorConfig, l *slog.Logger) *StoragePodSensor {
	var mc *sensormetrics.SensorMetricsCollector = sensormetrics.NewLastSensorMetricsCollector()
	var sm *sensormetrics.StatusMonitor = sensormetrics.NewStatusMonitor()
	return &StoragePodSensor{
		BaseSensor: *NewBaseSensor(
			"StoragePod", []string{
				"parent",
				"name",
				"summary",
			}, mc, sm),
		started:          helper.NewStartedCheck(),
		stopChan:         make(chan struct{}),
		manualRefresh:    make(chan struct{}),
		config:           config,
		SensorLogger:     logger.NewSLogLogger(l, logger.WithKind(STORAGE_POD_SENSOR_NAME)),
		metricsCollector: mc,
		statusMonitor:    sm,
	}
}

func (s *StoragePodSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	var spods []mo.StoragePod
	err := s.baseRefresh(ctx, scraper, &spods)
	if err != nil {
		return err
	}

	for _, spod := range spods {
		oSpod := ConvertToStoragePod(ctx, scraper, spod, time.Now())
		err := scraper.DB.SetStoragePod(ctx, oSpod, s.config.MaxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StoragePodSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *StoragePodSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
	ticker := time.NewTicker(s.config.RefreshInterval)
	go func() {
		time.Sleep(time.Duration(rand.Intn(20000)) * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				go func() {
					err := s.refresh(ctx, scraper)
					if err == nil {
						s.SensorLogger.Debug("refresh successful")
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

func (s *StoragePodSensor) StopRefresher(ctx context.Context) {
	close(s.stopChan)
}

func (s *StoragePodSensor) TriggerManualRefresh(ctx context.Context) {
	s.manualRefresh <- struct{}{}
}

func (s *StoragePodSensor) Kind() string {
	return "StoragePodSensor"
}

func (s *StoragePodSensor) WaitTillStartup() {
	s.started.Wait()
}

func (s *StoragePodSensor) Match(name string) bool {
	return helper.NewMatcher("storagepod", "storage_pod", "datastore_cluster", "datastorecluster").Match(name)
}

func (s *StoragePodSensor) Enabled() bool {
	return true
}

func (s *StoragePodSensor) GetLatestMetrics() []sensormetrics.SensorMetric {
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

func ConvertToStoragePod(ctx context.Context, scraper *VCenterScraper, p mo.StoragePod, t time.Time) objects.StoragePod {
	self := objects.NewManagedObjectReferenceFromVMwareRef(p.Self)

	var parent *objects.ManagedObjectReference
	if p.Parent != nil {
		p := objects.NewManagedObjectReferenceFromVMwareRef(*p.Parent)
		parent = &p
	}

	spod := objects.StoragePod{
		Timestamp: t,
		Name:      p.Name,
		Self:      self,
		Parent:    parent,
	}

	if spod.Parent != nil {
		parentChain := scraper.DB.GetParentChain(ctx, *spod.Parent)
		spod.Datacenter = parentChain.DC
	}

	if summary := p.Summary; summary != nil {
		spod.Capacity = float64(summary.Capacity)
		spod.FreeSpace = float64(summary.FreeSpace)
	}

	return spod
}
