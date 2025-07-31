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
	"github.com/vmware/govmomi/vim25/mo"
)

const STORAGE_POD_SENSOR_NAME = "StoragePodSensor"

type StoragePodSensor struct {
	BaseSensor
	logger.SensorLogger
	metricshelper.MetricHelperDefault
	started       helper.StartedCheck
	sensorLock    sync.Mutex
	manualRefresh chan struct{}
	stopChan      chan struct{}
	config        SensorConfig
}

func NewStoragePodSensor(scraper *VCenterScraper, config SensorConfig, l *slog.Logger) *StoragePodSensor {
	return &StoragePodSensor{
		BaseSensor: *NewBaseSensor(
			"StoragePod", []string{
				"parent",
				"name",
				"summary",
			}),
		stopChan:            make(chan struct{}),
		config:              config,
		SensorLogger:        logger.NewSLogLogger(l, logger.WithKind(STORAGE_POD_SENSOR_NAME)),
		MetricHelperDefault: *metricshelper.NewMetricHelperDefault(STORAGE_POD_SENSOR_NAME),
	}
}

func (s *StoragePodSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	var spods []mo.StoragePod
	stats, err := s.baseRefresh(ctx, scraper, &spods)
	s.MetricHelperDefault.LoadStats(stats)
	if err != nil {
		return err
	}

	for _, spod := range spods {
		oSpod := ConvertToStoragePod(ctx, scraper, spod, time.Now())
		err := scraper.DB.SetStoragePod(ctx, &oSpod, s.config.MaxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StoragePodSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) {
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
	return helper.NewMatcher("keyword").Match(name)
}

func (s *StoragePodSensor) Enabled() bool {
	return true
}

func ConvertToStoragePod(ctx context.Context, scraper *VCenterScraper, p mo.StoragePod, t time.Time) objects.StoragePod {
	self := objects.NewManagedObjectReference(p.Self.Type, p.Self.Value)

	var parent *objects.ManagedObjectReference
	if p.Parent != nil {
		p := objects.NewManagedObjectReference(parent.Type, parent.Value)
		parent = &p
	}

	spod := objects.StoragePod{
		Timestamp: t,
		Name:      p.Name,
		Self:      self,
		Parent:    parent,
	}

	if spod.Parent != nil {
		parentChain := scraper.GetParentChain(*spod.Parent)
		spod.Datacenter = parentChain.DC
	}

	if summary := p.Summary; summary != nil {
		spod.Capacity = float64(summary.Capacity)
		spod.FreeSpace = float64(summary.FreeSpace)
	}
	spod.OverallStatus = ConvertManagedEntityStatusToValue(p.OverallStatus)

	return spod
}
