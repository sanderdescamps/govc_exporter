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

type ComputeResourceSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.ComputeResource]
	AutoRunSensor
	Refreshable
	helper.Matchable
}

func NewComputeResourceSensor(scraper *VCenterScraper, config SensorConfig) *ComputeResourceSensor {
	var sensor ComputeResourceSensor
	sensor = ComputeResourceSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, mo.ComputeResource](
			scraper,
			"ComputeResourceSensor",
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

func (s *ComputeResourceSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
	if ok := s.sensorLock.TryLock(); !ok {
		logger.Info("Sensor Refresh already running", "sensor_type", s.Kind())
		return nil
	}
	defer s.sensorLock.Unlock()
	return s.unsafeRefresh(ctx, logger)
}

func (s *ComputeResourceSensor) unsafeRefresh(ctx context.Context, logger *slog.Logger) error {
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
		[]string{"ComputeResource"},
		true,
	)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	var computeResources []mo.ComputeResource
	err = v.Retrieve(
		context.Background(),
		[]string{"ComputeResource"},
		[]string{
			"parent",
			"summary",
			"name",
		},
		&computeResources,
	)
	t3 := time.Now()
	s.metrics.ClientWaitTime.Update(t2.Sub(t1))
	s.metrics.QueryTime.Update(t3.Sub(t2))
	s.metrics.Status.Update(true)
	if err != nil {
		s.metrics.Status.Update(true)
		return err
	}

	for _, computeResource := range computeResources {
		s.Update(computeResource.Self, &computeResource)
	}

	return nil
}

func (s *ComputeResourceSensor) Name() string {
	return "compute_resource"
}

func (s *ComputeResourceSensor) Match(name string) bool {
	return helper.NewMatcher("compute-resource", "computeresource").Match(name)
}
