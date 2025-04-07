package scraper

import (
	"context"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type DatastoreSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.Datastore]
	AutoRunSensor
	Refreshable
}

func NewDatastoreSensor(scraper *VCenterScraper, config SensorConfig) *DatastoreSensor {
	var sensor DatastoreSensor
	sensor = DatastoreSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, mo.Datastore](
			"datastore",
			"DatastoreSensor",
			helper.NewMatcher("datastore", "ds"),
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

func (s *DatastoreSensor) Refresh(ctx context.Context) error {
	entities, err := s.baseRefresh(ctx, "Datastore", []string{
		"name",
		"parent",
		"summary",
		"info",
	})
	if err != nil {
		return err
	}

	for _, entity := range entities {
		s.Update(entity.Self, &entity)
	}

	return nil
}
