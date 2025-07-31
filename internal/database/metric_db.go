package database

import (
	"context"
	"time"
)

type TimeItem interface {
	TimeKey() time.Time
}

type MetricDB interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Add(ctx context.Context, key string, path string, data ...TimeItem) error
	AddVmMetrics(ctx context.Context, VmID string, data ...TimeItem) error
	AddHostMetrics(ctx context.Context, HostID string, data ...TimeItem) error
	PopAll(ctx context.Context, key string, path string) []TimeItem
	PopOlderOrEqualThan(ctx context.Context, key string, path string, olderThan time.Time) []TimeItem
}
