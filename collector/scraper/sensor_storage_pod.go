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

type StoragePodSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.StoragePod]
	Refreshable
}

func NewStoragePodSensor(scraper *VCenterScraper) *StoragePodSensor {
	return &StoragePodSensor{
		BaseSensor: BaseSensor[types.ManagedObjectReference, mo.StoragePod]{
			cache:   make(map[types.ManagedObjectReference]*CacheItem[mo.StoragePod]),
			scraper: scraper,
			metrics: nil,
		},
	}
}

func (s *StoragePodSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
	sensorKind := reflect.TypeOf(s).String()
	if hasLock := s.refreshLock.TryLock(); hasLock {
		defer s.refreshLock.Unlock()
		return s.unsafeRefresh(ctx, logger)
	} else {
		logger.Info("Sensor Refresh already running", "sensor_type", sensorKind)
	}
	return nil
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
	s.setMetrics(&SensorMetric{
		Name:           "storage_pod",
		QueryTime:      t3.Sub(t2),
		ClientWaitTime: t2.Sub(t1),
		Status:         true,
	})
	if err != nil {
		return err
	}

	for _, storagePod := range storagePods {
		s.Update(storagePod.Self, &storagePod)
	}

	return nil
}
