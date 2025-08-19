package database

import (
	"context"
	"iter"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
)

type MetricDB interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error

	AddVmMetrics(ctx context.Context, ref objects.ManagedObjectReference, ttl time.Duration, data ...objects.Metric) error
	AddHostMetrics(ctx context.Context, ref objects.ManagedObjectReference, ttl time.Duration, data ...objects.Metric) error

	PopAllHostMetrics(ctx context.Context, ref objects.ManagedObjectReference) []*objects.Metric
	PopAllVmMetrics(ctx context.Context, ref objects.ManagedObjectReference) []*objects.Metric
	PopAllHostMetricsIter(ctx context.Context, ref objects.ManagedObjectReference) iter.Seq[objects.Metric]
	PopAllVmMetricsIter(ctx context.Context, ref objects.ManagedObjectReference) iter.Seq[objects.Metric]

	JsonDump(ctx context.Context, pmType ...objects.PerfMetricTypes) (map[objects.ManagedObjectReference][]byte, error)
}
