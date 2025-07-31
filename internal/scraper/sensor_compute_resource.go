package scraper

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper/logger"
	metricshelper "github.com/sanderdescamps/govc_exporter/internal/scraper/metrics_helper"
	"github.com/vmware/govmomi/vim25/mo"
)

const COMPUTE_RESOURCE_SENSOR_NAME = "ComputeResourceSensor"

type ComputeResourceSensor struct {
	BaseSensor
	logger.SensorLogger
	metricshelper.MetricHelperDefault
	started       helper.StartedCheck
	sensorLock    sync.Mutex
	manualRefresh chan struct{}
	stopChan      chan struct{}
	config        SensorConfig
}

func NewComputeResourceSensor(scraper *VCenterScraper, config SensorConfig, l *slog.Logger) *ComputeResourceSensor {
	return &ComputeResourceSensor{
		BaseSensor: *NewBaseSensor(
			"ComputeResource", []string{
				"parent",
				"name",
				"summary",
			}),
		stopChan:            make(chan struct{}),
		config:              config,
		SensorLogger:        logger.NewSLogLogger(l, logger.WithKind(COMPUTE_RESOURCE_SENSOR_NAME)),
		MetricHelperDefault: *metricshelper.NewMetricHelperDefault(COMPUTE_RESOURCE_SENSOR_NAME),
	}
}

func (s *ComputeResourceSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	var computeResources []mo.ComputeResource
	stats, err := s.baseRefresh(ctx, scraper, &computeResources)
	s.MetricHelperDefault.LoadStats(stats)
	if err != nil {
		return err
	}

	for _, cluster := range computeResources {
		err := scraper.DB.SetComputeResource(ctx, &cluster, s.config.MaxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ComputeResourceSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) {
	ticker := time.NewTicker(s.config.RefreshInterval)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				s.refresh(ctx, scraper)
				err := s.refresh(ctx, scraper)
				if err == nil {
					s.SensorLogger.Info("refresh successful")
				} else {
					s.SensorLogger.Error("refresh failed", "err", err)
				}
				s.started.Started()
			case <-s.manualRefresh:
				s.SensorLogger.Info("trigger manual refresh")
				err := s.refresh(ctx, scraper)
				if err == nil {
					s.SensorLogger.Info("manual refresh successful")
				} else {
					s.SensorLogger.Error("manual refresh failed", "err", err)
				}
			case <-s.stopChan:
				s.started.Stopped()
				return
			}
		}
	}()
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
	return helper.NewMatcher("keyword").Match(name)
}

func (s *ComputeResourceSensor) Enabled() bool {
	return true
}
