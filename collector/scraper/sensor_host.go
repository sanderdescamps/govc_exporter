package scraper

import (
	"context"
	"log/slog"
	"time"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type HostSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.HostSystem]
	Refreshable
	Cleanable
}

func NewHostSensor(scraper *VCenterScraper) *HostSensor {
	return &HostSensor{
		BaseSensor: BaseSensor[types.ManagedObjectReference, mo.HostSystem]{
			cache:   make(map[types.ManagedObjectReference]*CacheItem[mo.HostSystem]),
			scraper: scraper,
			metrics: nil,
		},
	}
}

func (s *HostSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
	t1 := time.Now()

	client, release, err := s.scraper.clientPool.Acquire()
	defer release()
	if err != nil {
		return err
	}
	t2 := time.Now()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"HostSystem"},
		true,
	)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	var items []mo.HostSystem
	err = v.Retrieve(
		context.Background(),
		[]string{"HostSystem"},
		[]string{
			"name",
			"parent",
			"summary",
			"runtime",
			"config.storageDevice",
			"config.fileSystemVolume",
			// "network",
		},
		&items,
	)
	t3 := time.Now()
	s.setMetrics(&SensorMetric{
		Name:           "host",
		QueryTime:      t3.Sub(t2),
		ClientWaitTime: t2.Sub(t1),
		Status:         true,
	})
	if err != nil {
		return err
	}

	for _, host := range items {
		s.Update(host.Self, &host)
	}

	return nil
}

func (s *HostSensor) Clean(maxAge time.Duration) {
	s.BaseSensor.Clean(maxAge)
}
