package scraper

import (
	"context"
	"log/slog"
	"reflect"
	"time"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type ComputeResourceSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.ComputeResource]
	Refreshable
}

func NewComputeResourceSensor(scraper *VCenterScraper) *ComputeResourceSensor {
	return &ComputeResourceSensor{
		BaseSensor: BaseSensor[types.ManagedObjectReference, mo.ComputeResource]{
			cache:   make(map[types.ManagedObjectReference]*CacheItem[mo.ComputeResource]),
			scraper: scraper,
			metrics: nil,
		},
	}
}

func (s *ComputeResourceSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
	sensorKind := reflect.TypeOf(s).String()
	if hasLock := s.refreshLock.TryLock(); hasLock {
		defer s.refreshLock.Unlock()
		return s.unsafeRefresh(ctx, logger)
	} else {
		logger.Info("Sensor Refresh already running", "sensor_type", sensorKind)
	}
	return nil
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
	s.setMetrics(&SensorMetric{
		Name:           "compute_resource",
		QueryTime:      t3.Sub(t2),
		ClientWaitTime: t2.Sub(t1),
		Status:         true,
	})
	if err != nil {
		return err
	}

	for _, computeResource := range computeResources {
		s.Update(computeResource.Self, &computeResource)
	}

	return nil
}
