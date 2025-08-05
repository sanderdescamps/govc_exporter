package memory_db

import (
	"context"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
)

type MetricsDB struct {
	queues map[objects.PerfMetricTypes]map[objects.ManagedObjectReference]*TimeQueueTable
	lock   sync.Mutex
}

func NewMetricsDB() *MetricsDB {
	return &MetricsDB{
		queues: make(map[objects.PerfMetricTypes]map[objects.ManagedObjectReference]*TimeQueueTable),
	}
}

func (db *MetricsDB) Connect(ctx context.Context) error {
	return nil
}

func (db *MetricsDB) Disconnect(ctx context.Context) error {
	db.queues = nil
	return nil
}

func (db *MetricsDB) Add(ctx context.Context, pmType objects.PerfMetricTypes, ref objects.ManagedObjectReference, data ...objects.Metric) error {
	db.Table(pmType, ref).Add(data...)
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

func (db *MetricsDB) AddVmMetrics(ctx context.Context, ref objects.ManagedObjectReference, data ...objects.Metric) error {
	return db.Add(ctx, objects.PerfMetricTypesVirtualMachine, ref, data...)
}

func (db *MetricsDB) AddHostMetrics(ctx context.Context, ref objects.ManagedObjectReference, data ...objects.Metric) error {
	return db.Add(ctx, objects.PerfMetricTypesHost, ref, data...)
}

func (db *MetricsDB) PopAll(ctx context.Context, pmType objects.PerfMetricTypes, ref objects.ManagedObjectReference) []*objects.Metric {
	return db.Table(pmType, ref).PopAll()
}

func (db *MetricsDB) PopOlderOrEqualThan(ctx context.Context, pmType objects.PerfMetricTypes, ref objects.ManagedObjectReference, olderThan time.Time) []*objects.Metric {
	return db.Table(pmType, ref).PopOlderOrEqualThan(olderThan)
}

func (db *MetricsDB) PopAllHostMetrics(ctx context.Context, ref objects.ManagedObjectReference) []*objects.Metric {
	return db.Table(objects.PerfMetricTypesHost, ref).PopAll()
}

func (db *MetricsDB) PopAllVmMetrics(ctx context.Context, ref objects.ManagedObjectReference) []*objects.Metric {
	return db.Table(objects.PerfMetricTypesVirtualMachine, ref).PopAll()
}
