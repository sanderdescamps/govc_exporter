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

const FOLDER_SENSOR_NAME = "Folder"

type FolderSensor struct {
	BaseSensor
	logger.SensorLogger
	metricshelper.MetricHelperDefault
	started       *helper.StartedCheck
	sensorLock    sync.Mutex
	manualRefresh chan struct{}
	stopChan      chan struct{}
	config        SensorConfig
}

func NewFolderSensor(scraper *VCenterScraper, config SensorConfig, l *slog.Logger) *FolderSensor {
	return &FolderSensor{
		BaseSensor: *NewBaseSensor(
			"Folder", []string{
				"parent",
				"name",
			}),
		started:             helper.NewStartedCheck(),
		stopChan:            make(chan struct{}),
		manualRefresh:       make(chan struct{}),
		config:              config,
		SensorLogger:        logger.NewSLogLogger(l, logger.WithKind(FOLDER_SENSOR_NAME)),
		MetricHelperDefault: *metricshelper.NewMetricHelperDefault(FOLDER_SENSOR_NAME),
	}
}

func (s *FolderSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	var folders []mo.Folder
	stats, err := s.baseRefresh(ctx, scraper, &folders)
	s.MetricHelperDefault.LoadStats(stats)
	if err != nil {
		return err
	}

	for _, folder := range folders {
		oFolder := ConvertToFolder(ctx, scraper, folder, time.Now())
		err := scraper.DB.SetFolder(ctx, oFolder, s.config.MaxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *FolderSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *FolderSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
	err := s.refresh(ctx, scraper)
	if err != nil {
		return err
	}

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

func (s *FolderSensor) StopRefresher(ctx context.Context) {
	close(s.stopChan)
}

func (s *FolderSensor) TriggerManualRefresh(ctx context.Context) {
	s.manualRefresh <- struct{}{}
}

func (s *FolderSensor) Kind() string {
	return FOLDER_SENSOR_NAME
}

func (s *FolderSensor) WaitTillStartup() {
	s.started.Wait()
}

func (s *FolderSensor) Match(name string) bool {
	return helper.NewMatcher("folder").Match(name)
}

func (s *FolderSensor) Enabled() bool {
	return true
}

func ConvertToFolder(ctx context.Context, scraper *VCenterScraper, f mo.Folder, t time.Time) objects.Folder {
	self := objects.NewManagedObjectReferenceFromVMwareRef(f.Self)

	var parent *objects.ManagedObjectReference
	if f.Parent != nil {
		p := objects.NewManagedObjectReferenceFromVMwareRef(*f.Parent)
		parent = &p
	}

	folder := objects.Folder{
		Timestamp: t,
		Name:      f.Name,
		Self:      self,
		Parent:    parent,
	}

	return folder
}
