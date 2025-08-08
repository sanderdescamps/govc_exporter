package redis_db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
)

type MetricsDB struct {
	client   *redis.Client
	password string
	addr     string
	dbIndex  int
	max_age  time.Duration
}

func NewMetricsDB(addr string, password string, index int) *MetricsDB {
	db := &MetricsDB{
		addr:     addr,
		password: password,
		dbIndex:  index,
	}
	return db
}

func (db *MetricsDB) Connect(ctx context.Context) error {
	if db.client != nil {
		return nil
	}

	db.client = redis.NewClient(&redis.Options{
		Addr:     db.addr,
		Password: db.password, // no password set
		DB:       db.dbIndex,  // use default DB
	})

	return nil
}

func (db *MetricsDB) Disconnect(ctx context.Context) error {
	db.client = nil
	return nil
}

func (db *MetricsDB) Add(ctx context.Context, pmType objects.PerfMetricTypes, ref objects.ManagedObjectReference, ttl time.Duration, data ...objects.Metric) error {
	db.Connect(ctx)

	for _, metric := range data {
		redisKey := fmt.Sprintf("%s:%s:%s:%s", pmType.String(), ref.Type.String(), ref.Value, metric.Hash())
		status := db.client.JSONSet(ctx, redisKey, "$", metric)
		if status.Err() != nil {
			return status.Err()
		}

		if ttl != 0 {
			_, err := db.client.Expire(ctx, redisKey, ttl).Result()
			if err != nil {
				return fmt.Errorf("failed to set expiration of object: %v", err)
			}
		}
	}

	return nil
}

func (db *MetricsDB) AddVmMetrics(ctx context.Context, ref objects.ManagedObjectReference, ttl time.Duration, data ...objects.Metric) error {
	return db.Add(ctx, objects.PerfMetricTypesVirtualMachine, ref, ttl, data...)
}

func (db *MetricsDB) AddHostMetrics(ctx context.Context, ref objects.ManagedObjectReference, ttl time.Duration, data ...objects.Metric) error {
	return db.Add(ctx, objects.PerfMetricTypesHost, ref, ttl, data...)
}

func (db *MetricsDB) PopAll(ctx context.Context, pmType objects.PerfMetricTypes, ref objects.ManagedObjectReference) []*objects.Metric {
	db.Connect(ctx)
	result := []*objects.Metric{}
	match := fmt.Sprintf("%s:%s:%s:*", pmType.String(), ref.Type.String(), ref.Value)
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	for redisIter.Next(ctx) {
		redisKey := redisIter.Val()
		redisCmd := db.client.JSONGet(ctx, redisKey, ".")
		if redisCmd.Err() != nil {
			return nil
		}

		status := db.client.Del(ctx, redisKey)
		if status.Err() != nil {
			return nil
		}

		var metric objects.Metric
		if err := json.Unmarshal([]byte(redisCmd.Val()), &metric); err != nil {
			panic("failed to parse metric")
		}
		result = append(result, &metric)
	}

	return result
}

func (db *MetricsDB) PopOlderOrEqualThan(ctx context.Context, pmType objects.PerfMetricTypes, ref objects.ManagedObjectReference, olderThan time.Time) []*objects.Metric {
	return []*objects.Metric{}
}

func (db *MetricsDB) PopAllHostMetrics(ctx context.Context, ref objects.ManagedObjectReference) []*objects.Metric {
	return db.PopAll(ctx, objects.PerfMetricTypesHost, ref)
}

func (db *MetricsDB) PopAllVmMetrics(ctx context.Context, ref objects.ManagedObjectReference) []*objects.Metric {
	return db.PopAll(ctx, objects.PerfMetricTypesVirtualMachine, ref)
}
