package scraper

import (
	"context"
	"log/slog"
	"time"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type DatastoreSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.Datastore]
	Refreshable
	Cleanable
}

func NewDatastoreSensor(scraper *VCenterScraper) *DatastoreSensor {
	return &DatastoreSensor{
		BaseSensor: BaseSensor[types.ManagedObjectReference, mo.Datastore]{
			cache:   make(map[types.ManagedObjectReference]*CacheItem[mo.Datastore]),
			scraper: scraper,
			metrics: nil,
		},
	}
}

func (s *DatastoreSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
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
		[]string{"Datastore"},
		true,
	)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	var datastores []mo.Datastore
	err = v.Retrieve(
		ctx,
		[]string{"Datastore"},
		[]string{
			"name",
			"parent",
			"summary",
			"info",
		},
		&datastores,
	)
	t3 := time.Now()
	s.setMetrics(&SensorMetric{
		Name:           "datastore",
		QueryTime:      t3.Sub(t2),
		ClientWaitTime: t2.Sub(t1),
		Status:         true,
	})
	if err != nil {
		return err
	}

	for _, datastore := range datastores {
		s.Update(datastore.Self, &datastore)
	}

	return nil
}

func (s *DatastoreSensor) Clean(maxAge time.Duration, logger *slog.Logger) {
	s.BaseSensor.Clean(maxAge, logger)
}
