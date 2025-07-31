package memory_db

import (
	"context"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database"
)

type MetricsDB struct {
	queues map[string]map[string]*TimeQueueTable
}

func NewMetricsDB() *MetricsDB {
	return &MetricsDB{
		queues: make(map[string]map[string]*TimeQueueTable),
	}
}

func (db *MetricsDB) Connect(ctx context.Context) error {
	return nil
}

func (db *MetricsDB) Disconnect(ctx context.Context) error {
	db.queues = nil
	return nil
}

func (db *MetricsDB) Add(ctx context.Context, key string, path string, data ...database.TimeItem) error {
	if _, ok := db.queues[path]; !ok {
		db.queues[path] = make(map[string]*TimeQueueTable)
	}
	if _, ok := db.queues[path][key]; !ok {
		db.queues[path][key] = NewTimeQueueTable()
	}

	db.queues[path][key].Add(data...)
	return nil
}

func (db *MetricsDB) AddVmMetrics(ctx context.Context, key string, data ...database.TimeItem) error {
	return db.Add(ctx, key, "VirtualMachine", data...)
}

func (db *MetricsDB) AddHostMetrics(ctx context.Context, key string, data ...database.TimeItem) error {
	return db.Add(ctx, key, "HostSystem", data...)
}

func (db *MetricsDB) PopAll(ctx context.Context, key string, path string) []database.TimeItem {
	if queue, ok := db.queues[path][key]; ok {
		return queue.PopAll()
	}
	return nil
}

func (db *MetricsDB) PopOlderOrEqualThan(ctx context.Context, key string, path string, olderThan time.Time) []database.TimeItem {
	if queue, ok := db.queues[path][key]; ok {
		return queue.PopOlderOrEqualThan(olderThan)
	}
	return nil
}
