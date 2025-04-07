package scraper

import (
	"context"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type StoragePodSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.StoragePod]
	AutoRunSensor
	Refreshable
}

func NewStoragePodSensor(scraper *VCenterScraper, config SensorConfig) *StoragePodSensor {
	var sensor StoragePodSensor
	sensor = StoragePodSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, mo.StoragePod](
			"storage_pod",
			"StoragePodSensor",
			helper.NewMatcher("storagepod", "storage_pod", "datastore_cluster", "datastorecluster"),
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

func (s *StoragePodSensor) Refresh(ctx context.Context) error {
	entities, err := s.baseRefresh(ctx, "StoragePod", []string{
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
