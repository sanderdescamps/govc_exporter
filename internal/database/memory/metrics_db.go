package memory_db

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database"
	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
)

type metricIndex struct {
	pmType objects.PerfMetricTypes
	ref    objects.ManagedObjectReference
}

type MetricsDB struct {
	tables          map[metricIndex]*TimeQueueTable
	lock            sync.Mutex
	cleanerStopChan chan bool
}

func NewMetricsDB() *MetricsDB {
	return &MetricsDB{
		tables: make(map[metricIndex]*TimeQueueTable),
	}
}

func (db *MetricsDB) Connect(ctx context.Context) error {
	db.startCleaner(ctx)
	return nil
}

func (db *MetricsDB) Disconnect(ctx context.Context) error {
	db.stopCleaner(ctx)
	db.tables = nil
	return nil
}

func (db *MetricsDB) startCleaner(ctx context.Context) {
	ticker := time.NewTicker(CLEANUP_INTERVAL)
	go func() {
		for {
			select {
			case <-ticker.C:
				go func() {
					counts := map[objects.PerfMetricTypes]int{}
					for index, table := range maps.All(db.tables) {
						counts[index.pmType] += table.CleanupExpired()
					}

					for pmType, count := range counts {
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
	db.lock.Lock()
	defer db.lock.Unlock()

	index := metricIndex{pmType: pmType, ref: ref}
	if _, ok := db.tables[index]; !ok {
		db.tables[index] = NewTimeQueueTable()
	}
	return db.tables[index]
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
func (db *MetricsDB) JsonDump(ctx context.Context, pmType ...objects.PerfMetricTypes) (map[objects.ManagedObjectReference][]byte, error) {
	db.lock.Lock()
	defer db.lock.Unlock()
	result := make(map[objects.ManagedObjectReference][]byte)
	for index, table := range db.tables {
		if helper.Contains(pmType, index.pmType) {
			b, err := table.JsonDump()
			if err != nil {
				return nil, err
			}
			result[index.ref] = b
		}
	}

	return result, nil
}
