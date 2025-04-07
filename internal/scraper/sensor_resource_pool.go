package scraper

import (
	"context"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type ResourcePoolSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.ResourcePool]
	AutoRunSensor
	Refreshable
}

func NewResourcePoolSensor(scraper *VCenterScraper, config SensorConfig) *ResourcePoolSensor {
	var sensor ResourcePoolSensor
	sensor = ResourcePoolSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, mo.ResourcePool](
			"resource_pool",
			"ResourcePoolSensor",
			helper.NewMatcher("resource_pool", "resourcepool", "repool"),
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

func (s *ResourcePoolSensor) Refresh(ctx context.Context) error {
	entities, err := s.baseRefresh(ctx, "ResourcePool", []string{
		"parent",
		"summary",
		"name",
	})
	if err != nil {
		return err
	}

	for _, entity := range entities {
		s.Update(entity.Self, &entity)
	}

	return nil
}
