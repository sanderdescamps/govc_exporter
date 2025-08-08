package database

import (
	"context"
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
}
