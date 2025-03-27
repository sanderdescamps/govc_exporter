package scraper

import (
	"context"
	"log/slog"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type StoragePodSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.StoragePod]
	AutoRunSensor
	Refreshable
	helper.Matchable
}

func NewStoragePodSensor(scraper *VCenterScraper, config SensorConfig) *StoragePodSensor {
	var sensor StoragePodSensor
	sensor = StoragePodSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, mo.StoragePod](
			scraper,
			"StoragePodSensor",
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

func (s *StoragePodSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
	if ok := s.sensorLock.TryLock(); !ok {
		logger.Info("Sensor Refresh already running", "sensor_type", s.Kind())
		return nil
	}
	defer s.sensorLock.Unlock()
	return s.unsafeRefresh(ctx, logger)
}

func (s *StoragePodSensor) unsafeRefresh(ctx context.Context, logger *slog.Logger) error {
	t1 := time.Now()
	client, release, err := s.scraper.clientPool.Acquire()
	if err != nil {
		return err
	}
	defer release()
	t2 := time.Now()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"StoragePod"},
		true,
	)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	var storagePods []mo.StoragePod
	err = v.Retrieve(
		ctx,
		[]string{"StoragePod"},
		[]string{
			"name",
			"parent",
			"summary",
		},
		&storagePods,
	)
	t3 := time.Now()
	s.metrics.ClientWaitTime.Update(t2.Sub(t1))
	s.metrics.QueryTime.Update(t3.Sub(t2))
	s.metrics.Status.Update(true)
	if err != nil {
		s.metrics.Status.Update(true)
		return err
	}

	for _, storagePod := range storagePods {
		s.Update(storagePod.Self, &storagePod)
	}

	return nil
}

func (s *StoragePodSensor) Name() string {
	return "storage_pod"
}

func (s *StoragePodSensor) Match(name string) bool {
	return helper.NewMatcher("storagepod", "storage_pod", "datastore_cluster", "datastorecluster").Match(name)
}
