package scraper

import (
	"context"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type ComputeResourceSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.ComputeResource]
	AutoRunSensor
	Refreshable
}

func NewComputeResourceSensor(scraper *VCenterScraper, config SensorConfig) *ComputeResourceSensor {
	var sensor ComputeResourceSensor
	sensor = ComputeResourceSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, mo.ComputeResource](
			"compute_resource",
			"ComputeResourceSensor",
			helper.NewMatcher("compute-resource", "computeresource"),
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

func (s *ComputeResourceSensor) Refresh(ctx context.Context) error {
	entities, err := s.baseRefresh(ctx, "ComputeResource", []string{
		"parent",
		"name",
		"summary",
	})
	if err != nil {
		return err
	}

	for _, entity := range entities {
		s.Update(entity.Self, &entity)
	}

	return nil

}
