package scraper

import (
	"context"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type HostSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.HostSystem]
	AutoRunSensor
	Refreshable
}

func NewHostSensor(scraper *VCenterScraper, config SensorConfig) *HostSensor {
	var sensor HostSensor
	sensor = HostSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, mo.HostSystem](
			"host",
			"HostSensor",
			helper.NewMatcher("host", "esx"),
			scraper,
		),
		AutoRunSensor: *NewAutoRunSensor(&sensor, config),
	}
	sensor.metrics.ClientWaitTime = NewSensorMetricDuration(sensor.Kind(), "client_wait_time", 0)
	sensor.metrics.QueryTime = NewSensorMetricDuration(sensor.Kind(), "query_time", 0)
	sensor.metrics.Status = NewSensorMetricStatus(sensor.Kind(), "status", false)
	scraper.RegisterSensorMetric(
		&sensor.metrics.ClientWaitTime.SensorMetric,
		&sensor.metrics.QueryTime.SensorMetric,
		&sensor.metrics.Status.SensorMetric,
	)
	return &sensor
}

func (s *HostSensor) Refresh(ctx context.Context) error {
	entities, err := s.baseRefresh(ctx, "HostSystem", []string{
		"name",
		"parent",
		"summary",
		"runtime",
		"config.storageDevice",
		"config.fileSystemVolume",
		// "network",
	})
	if err != nil {
		return err
	}

	for _, entity := range entities {
		// set some unused parts of the object to nil to reduse memory usage
		entity.Summary.Hardware.OtherIdentifyingInfo = nil
		entity.Summary.Runtime = nil
		entity.Config = nil
		s.Update(entity.Self, &entity)
	}

	return nil
}
