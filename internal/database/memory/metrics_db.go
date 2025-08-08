package memory_db

import (
	"context"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
)

type MetricsDB struct {
	queues          map[objects.PerfMetricTypes]map[objects.ManagedObjectReference]*TimeQueueTable
	lock            sync.Mutex
	cleanerStopChan chan bool
}

func NewMetricsDB() *MetricsDB {
	return &MetricsDB{
		queues: make(map[objects.PerfMetricTypes]map[objects.ManagedObjectReference]*TimeQueueTable),
	}
}

func (db *MetricsDB) Connect(ctx context.Context) error {
	db.startCleaner(ctx)
	return nil
}

func (db *MetricsDB) Disconnect(ctx context.Context) error {
	db.stopCleaner(ctx)
	db.queues = nil
	return nil
}

func (db *MetricsDB) startCleaner(ctx context.Context) {
	ticker := time.NewTicker(CLEANUP_INTERVAL)
	go func() {
		for {
			select {
			case <-ticker.C:
				go func() {
					for pmType := range db.queues {
						for ref := range db.queues[pmType] {
							if q := db.Table(pmType, ref); q != nil {
								q.CleanupExpired()
							}
						}
					}
				}()
			case <-db.cleanerStopChan:
				ticker.Stop()
				return
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (db *MetricsDB) stopCleaner(ctx context.Context) {
	db.cleanerStopChan <- true
}

func (db *MetricsDB) Add(ctx context.Context, pmType objects.PerfMetricTypes, ref objects.ManagedObjectReference, ttl time.Duration, data ...objects.Metric) error {
	db.Table(pmType, ref).Add(ttl, data...)
	return nil
}

func (db *MetricsDB) Table(pmType objects.PerfMetricTypes, ref objects.ManagedObjectReference) *TimeQueueTable {
	db.lock.Lock()
	defer db.lock.Unlock()
	if _, ok := db.queues[pmType]; !ok {
		db.queues[pmType] = make(map[objects.ManagedObjectReference]*TimeQueueTable)
	}
	if _, ok := db.queues[pmType][ref]; !ok {
		db.queues[pmType][ref] = NewTimeQueueTable()
	}
	return db.queues[pmType][ref]
}

func (db *MetricsDB) AddVmMetrics(ctx context.Context, ref objects.ManagedObjectReference, ttl time.Duration, data ...objects.Metric) error {
	return db.Add(ctx, objects.PerfMetricTypesVirtualMachine, ref, ttl, data...)
}

func (db *MetricsDB) AddHostMetrics(ctx context.Context, ref objects.ManagedObjectReference, ttl time.Duration, data ...objects.Metric) error {
	return db.Add(ctx, objects.PerfMetricTypesHost, ref, ttl, data...)
}

func (db *MetricsDB) PopAll(ctx context.Context, pmType objects.PerfMetricTypes, ref objects.ManagedObjectReference) []*objects.Metric {
	return db.Table(pmType, ref).PopAll()
}

func (db *MetricsDB) PopAllHostMetrics(ctx context.Context, ref objects.ManagedObjectReference) []*objects.Metric {
	return db.Table(objects.PerfMetricTypesHost, ref).PopAll()
}

func (db *MetricsDB) PopAllVmMetrics(ctx context.Context, ref objects.ManagedObjectReference) []*objects.Metric {
	return db.Table(objects.PerfMetricTypesVirtualMachine, ref).PopAll()
}
