package scraper

import (
	"context"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type ClusterSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.ClusterComputeResource]
	AutoRunSensor
	Refreshable
}

func NewClusterSensor(scraper *VCenterScraper, config SensorConfig) *ClusterSensor {
	var sensor ClusterSensor
	sensor = ClusterSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, mo.ClusterComputeResource](
			"cluster",
			"ClusterSensor",
			helper.NewMatcher("cluster"),
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

func (s *ClusterSensor) Refresh(ctx context.Context) error {
	entities, err := s.baseRefresh(ctx, "ClusterComputeResource", []string{
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
