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

const DATACENTER_SENSOR_NAME = "Datacenter"

type DatacenterSensor struct {
	BaseSensor
	logger.SensorLogger
	metricshelper.MetricHelperDefault
	started       *helper.StartedCheck
	sensorLock    sync.Mutex
	manualRefresh chan struct{}
	stopChan      chan struct{}
	config        SensorConfig
}

func NewDatacenterSensor(scraper *VCenterScraper, config SensorConfig, l *slog.Logger) *DatacenterSensor {
	return &DatacenterSensor{
		BaseSensor: *NewBaseSensor(
			"Datacenter", []string{
				"parent",
				"name",
			}),
		started:             helper.NewStartedCheck(),
		stopChan:            make(chan struct{}),
		manualRefresh:       make(chan struct{}),
		config:              config,
		SensorLogger:        logger.NewSLogLogger(l, logger.WithKind(DATACENTER_SENSOR_NAME)),
		MetricHelperDefault: *metricshelper.NewMetricHelperDefault(DATACENTER_SENSOR_NAME),
	}
}

func (s *DatacenterSensor) refresh(ctx context.Context, scraper *VCenterScraper) error {
	if ok := s.sensorLock.TryLock(); !ok {
		return ErrSensorAlreadyRunning
	}
	defer s.sensorLock.Unlock()

	var datacenters []mo.Datacenter
	stats, err := s.baseRefresh(ctx, scraper, &datacenters)
	s.MetricHelperDefault.LoadStats(stats)
	if err != nil {
		return err
	}

	for _, dc := range datacenters {
		oDC := ConvertToDatacenter(ctx, scraper, dc, time.Now())
		err := scraper.DB.SetDatacenter(ctx, oDC, s.config.MaxAge)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *DatacenterSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *DatacenterSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
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

func (s *DatacenterSensor) StopRefresher(ctx context.Context) {
	close(s.stopChan)
}

func (s *DatacenterSensor) TriggerManualRefresh(ctx context.Context) {
	s.manualRefresh <- struct{}{}
}

func (s *DatacenterSensor) Kind() string {
	return DATACENTER_SENSOR_NAME
}

func (s *DatacenterSensor) WaitTillStartup() {
	s.started.Wait()
}

func (s *DatacenterSensor) Match(name string) bool {
	return helper.NewMatcher("datacenter", "dc").Match(name)
}

func (s *DatacenterSensor) Enabled() bool {
	return true
}

func ConvertToDatacenter(ctx context.Context, scraper *VCenterScraper, dc mo.Datacenter, t time.Time) objects.Datacenter {
	self := objects.NewManagedObjectReferenceFromVMwareRef(dc.Self)

	var parent *objects.ManagedObjectReference
	if dc.Parent != nil {
		p := objects.NewManagedObjectReferenceFromVMwareRef(*dc.Parent)
		parent = &p
	}

	datacenter := objects.Datacenter{
		Timestamp: t,
		Name:      dc.Name,
		Self:      self,
		Parent:    parent,
	}

	return datacenter
}
