package scraper

import (
	"context"
	"log/slog"
	"time"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type ClusterSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.ClusterComputeResource]
}

func NewClusterSensor(scraper *VCenterScraper) *ClusterSensor {
	return &ClusterSensor{
		BaseSensor: BaseSensor[types.ManagedObjectReference, mo.ClusterComputeResource]{
			cache:   make(map[types.ManagedObjectReference]*CacheItem[mo.ClusterComputeResource]),
			scraper: scraper,
			metrics: nil,
		},
	}
}

func (s *ClusterSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
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
	s.setMetrics(&SensorMetric{
		Name:           "cluster",
		QueryTime:      t3.Sub(t2),
		ClientWaitTime: t2.Sub(t1),
		Status:         true,
	})
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		s.Update(cluster.Self, &cluster)
	}

	return nil
}

func (s *ClusterSensor) Clean(maxAge time.Duration, logger *slog.Logger) {
	s.BaseSensor.Clean(maxAge, logger)
}
