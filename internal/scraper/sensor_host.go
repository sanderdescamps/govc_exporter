package scraper

import (
	"context"
	"log/slog"
	"reflect"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type HostSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.HostSystem]
	Refreshable
	helper.Matchable
}

func NewHostSensor(scraper *VCenterScraper) *HostSensor {
	sensor := &HostSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, mo.HostSystem](
			scraper,
		),
	}
	sensor.metrics.ClientWaitTime = NewSensorMetricDuration("sensor.host.client_wait_time", 0)
	sensor.metrics.QueryTime = NewSensorMetricDuration("sensor.host.query_time", 0)
	sensor.metrics.Status = NewSensorMetricStatus("sensor.host.status", false)
	scraper.RegisterSensorMetric(
		&sensor.metrics.ClientWaitTime.SensorMetric,
		&sensor.metrics.QueryTime.SensorMetric,
		&sensor.metrics.Status.SensorMetric,
	)
	return sensor
}

func (s *HostSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
	sensorKind := reflect.TypeOf(s).String()
	if hasLock := s.refreshLock.TryLock(); hasLock {
		defer s.refreshLock.Unlock()
		return s.unsafeRefresh(ctx, logger)
	} else {
		logger.Info("Sensor Refresh already running", "sensor_type", sensorKind)
	}
	return nil
}

func (s *HostSensor) unsafeRefresh(ctx context.Context, logger *slog.Logger) error {
	t1 := time.Now()

	client, release, err := s.scraper.clientPool.Acquire()
	defer release()
	if err != nil {
		return err
	}
	t2 := time.Now()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"HostSystem"},
		true,
	)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	var items []mo.HostSystem
	err = v.Retrieve(
		context.Background(),
		[]string{"HostSystem"},
		[]string{
			"name",
			"parent",
			"summary",
			"runtime",
			"config.storageDevice",
			"config.fileSystemVolume",
			// "network",
		},
		&items,
	)
	t3 := time.Now()
	s.metrics.ClientWaitTime.Update(t2.Sub(t1))
	s.metrics.QueryTime.Update(t3.Sub(t2))
	s.metrics.Status.Update(true)
	if err != nil {
		s.metrics.Status.Update(true)
		return err
	}

	for _, host := range items {
		s.Update(host.Self, &host)
	}

	return nil
}

func (s *HostSensor) Name() string {
	return "host"
}

func (s *HostSensor) Match(name string) bool {
	return helper.NewMatcher("host", "esx").Match(name)
}
