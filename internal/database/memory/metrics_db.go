package memory_db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database"
	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
)

type MetricsDB struct {
	queues          map[objects.PerfMetricTypes]map[string]*TimeQueueTable
	lock            sync.Mutex
	cleanerStopChan chan bool
}

func NewMetricsDB() *MetricsDB {
	return &MetricsDB{
		queues: make(map[objects.PerfMetricTypes]map[string]*TimeQueueTable),
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
						count := 0
						for ref := range db.queues[pmType] {
							if q := db.tableByRefID(pmType, ref); q != nil {
								count += q.CleanupExpired()
							}
						}
						if l := database.GetLoggerFromContext(ctx); l != nil && count > 0 {
							l.Info(fmt.Sprintf("removed %d metrics from %s table", count, pmType.String()))
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
	return db.tableByRefID(pmType, ref.ID())
}

func (db *MetricsDB) tableByRefID(pmType objects.PerfMetricTypes, refID string) *TimeQueueTable {
	db.lock.Lock()
	defer db.lock.Unlock()
	if _, ok := db.queues[pmType]; !ok {
		db.queues[pmType] = make(map[string]*TimeQueueTable)
	}
	if _, ok := db.queues[pmType][refID]; !ok {
		db.queues[pmType][refID] = NewTimeQueueTable()
	}
	return db.queues[pmType][refID]
}

func (db *MetricsDB) AddVmMetrics(ctx context.Context, ref objects.ManagedObjectReference, ttl time.Duration, data ...objects.Metric) error {
	return db.Add(ctx, objects.PerfMetricTypesVirtualMachine, ref, ttl, data...)
}

func (db *MetricsDB) AddHostMetrics(ctx context.Context, ref objects.ManagedObjectReference, ttl time.Duration, data ...objects.Metric) error {
	return db.Add(ctx, objects.PerfMetricTypesHost, ref, ttl, data...)
}

func (db *MetricsDB) PopAll(ctx context.Context, pmType objects.PerfMetricTypes, ref objects.ManagedObjectReference) []*objects.Metric {
	return db.Table(pmType, ref).PopAll()

	// return db.Table(pmType, ref).PopAll()
}

func (db *MetricsDB) PopAllHostMetrics(ctx context.Context, ref objects.ManagedObjectReference) []*objects.Metric {
	return db.Table(objects.PerfMetricTypesHost, ref).PopAll()
}

func (db *MetricsDB) PopAllVmMetrics(ctx context.Context, ref objects.ManagedObjectReference) []*objects.Metric {
	return db.Table(objects.PerfMetricTypesVirtualMachine, ref).PopAll()
}
