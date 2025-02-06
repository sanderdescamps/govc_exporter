package scraper

import (
	"context"
	"log/slog"
	"time"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type ResourcePoolSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.ResourcePool]
	Refreshable
	Cleanable
}

func NewResourcePoolSensor(scraper *VCenterScraper) *ResourcePoolSensor {
	return &ResourcePoolSensor{
		BaseSensor: BaseSensor[types.ManagedObjectReference, mo.ResourcePool]{
			cache:   make(map[types.ManagedObjectReference]*CacheItem[mo.ResourcePool]),
			scraper: scraper,
			metrics: nil,
		},
	}
}

func (s *ResourcePoolSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
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
		[]string{"ResourcePool"},
		true,
	)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	var resourcePools []mo.ResourcePool
	err = v.Retrieve(
		context.Background(),
		[]string{"ResourcePool"},
		[]string{
			"parent",
			"summary",
			"name",
		},
		&resourcePools,
	)
	t3 := time.Now()
	s.setMetrics(&SensorMetric{
		Name:           "resource_pool",
		QueryTime:      t3.Sub(t2),
		ClientWaitTime: t2.Sub(t1),
		Status:         true,
	})
	if err != nil {
		return err
	}

	for _, resourcePool := range resourcePools {
		s.Update(resourcePool.Self, &resourcePool)
	}

	return nil
}

func (s *ResourcePoolSensor) Clean(maxAge time.Duration, logger *slog.Logger) {
	s.BaseSensor.Clean(maxAge, logger)
}
