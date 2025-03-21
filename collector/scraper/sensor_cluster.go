package scraper

import (
	"context"
	"log/slog"
	"reflect"
	"time"

	"github.com/sanderdescamps/govc_exporter/collector/helper"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type ClusterSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.ClusterComputeResource]
	Refreshable
	helper.Matchable
}

func NewClusterSensor(scraper *VCenterScraper) *ClusterSensor {
	sensor := &ClusterSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, mo.ClusterComputeResource](
			scraper,
		),
	}
	sensor.metrics.ClientWaitTime = NewSensorMetricDuration("sensor.cluster.client_wait_time", 0)
	sensor.metrics.QueryTime = NewSensorMetricDuration("sensor.cluster.query_time", 0)
	sensor.metrics.Status = NewSensorMetricStatus("sensor.cluster.status", false)
	scraper.RegisterSensorMetric(
		&sensor.metrics.ClientWaitTime.SensorMetric,
		&sensor.metrics.QueryTime.SensorMetric,
		&sensor.metrics.Status.SensorMetric,
	)
	return sensor
}

func (s *ClusterSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
	sensorKind := reflect.TypeOf(s).String()
	if hasLock := s.refreshLock.TryLock(); hasLock {
		defer s.refreshLock.Unlock()
		return s.unsafeRefresh(ctx, logger)
	} else {
		logger.Info("Sensor Refresh already running", "sensor_type", sensorKind)
	}
	return nil
}

func (s *ClusterSensor) unsafeRefresh(ctx context.Context, logger *slog.Logger) error {
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
		[]string{"ClusterComputeResource"},
		true,
	)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	var clusters []mo.ClusterComputeResource
	err = v.Retrieve(
		context.Background(),
		[]string{"ClusterComputeResource"},
		[]string{
			"parent",
			"name",
			"summary",
		},
		&clusters,
	)
	t3 := time.Now()
	s.metrics.ClientWaitTime.Update(t2.Sub(t1))
	s.metrics.QueryTime.Update(t3.Sub(t2))
	s.metrics.Status.Update(true)
	if err != nil {
		s.metrics.Status.Update(true)
		return err
	}

	for _, cluster := range clusters {
		s.Update(cluster.Self, &cluster)
	}

	return nil
}

func (s *ClusterSensor) Name() string {
	return "cluster"
}

func (s *ClusterSensor) Match(name string) bool {
	return helper.NewMatcher("cluster").Match(name)
}
