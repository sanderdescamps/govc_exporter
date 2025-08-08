package scraper

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper/logger"
	sensormetrics "github.com/sanderdescamps/govc_exporter/internal/scraper/sensor_metrics"
	"github.com/vmware/govmomi/vim25/mo"
)

const COMPUTE_RESOURCE_SENSOR_NAME = "ComputeResourceSensor"

type ComputeResourceSensor struct {
	BaseSensor
	logger.SensorLogger
	metricsCollector *sensormetrics.SensorMetricsCollector
	statusMonitor    *sensormetrics.StatusMonitor
	started          *helper.StartedCheck
	sensorLock       sync.Mutex
	manualRefresh    chan struct{}
	stopChan         chan struct{}
	config           SensorConfig
}

func NewComputeResourceSensor(scraper *VCenterScraper, config SensorConfig, l *slog.Logger) *ComputeResourceSensor {
	var mc *sensormetrics.SensorMetricsCollector = sensormetrics.NewLastSensorMetricsCollector()
	var sm *sensormetrics.StatusMonitor = sensormetrics.NewStatusMonitor()
	return &ComputeResourceSensor{
		BaseSensor: *NewBaseSensor(
			"ComputeResource", []string{
				"parent",
				"name",
				"summary",
			}, mc, sm),
		started:          helper.NewStartedCheck(),
		stopChan:         make(chan struct{}),
		manualRefresh:    make(chan struct{}),
		config:           config,
		SensorLogger:     logger.NewSLogLogger(l, logger.WithKind(COMPUTE_RESOURCE_SENSOR_NAME)),
		metricsCollector: mc,
		statusMonitor:    sm,
	}
}

func (s *ComputeResourceSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	var computeResources []mo.ComputeResource
	err := s.baseRefresh(ctx, scraper, &computeResources)
	if err != nil {
		return err
	}

	for _, compResource := range computeResources {
		oCompResource := ConvertToComputeResource(ctx, scraper, compResource, time.Now())
		err := scraper.DB.SetComputeResource(ctx, oCompResource, s.config.MaxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ComputeResourceSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *ComputeResourceSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *ComputeResourceSensor) StopRefresher(ctx context.Context) {
	close(s.stopChan)
}

func (s *ComputeResourceSensor) TriggerManualRefresh(ctx context.Context) {
	s.manualRefresh <- struct{}{}
}

func (s *ComputeResourceSensor) Kind() string {
	return "ComputeResourceSensor"
}

func (s *ComputeResourceSensor) WaitTillStartup() {
	s.started.Wait()
}

func (s *ComputeResourceSensor) Match(name string) bool {
	return helper.NewMatcher("compute-resource", "computeresource").Match(name)
}

func (s *ComputeResourceSensor) Enabled() bool {
	return true
}

func (s *ComputeResourceSensor) GetLatestMetrics() []sensormetrics.SensorMetric {
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

func ConvertToComputeResource(ctx context.Context, scraper *VCenterScraper, r mo.ComputeResource, t time.Time) objects.ComputeResource {
	self := objects.NewManagedObjectReferenceFromVMwareRef(r.Self)

	var parent *objects.ManagedObjectReference
	if r.Parent != nil {
		p := objects.NewManagedObjectReferenceFromVMwareRef(*r.Parent)
		parent = &p
	}

	computeResource := objects.ComputeResource{
		Timestamp: t,
		Name:      r.Name,
		Self:      self,
		Parent:    parent,
	}

	return computeResource
}
