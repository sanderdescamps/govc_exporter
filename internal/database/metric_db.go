package database

import (
	"context"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
)

type MetricDB interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error

	AddVmMetrics(ctx context.Context, ref objects.ManagedObjectReference, data ...objects.Metric) error
	AddHostMetrics(ctx context.Context, ref objects.ManagedObjectReference, data ...objects.Metric) error

	PopAllHostMetrics(ctx context.Context, ref objects.ManagedObjectReference) []*objects.Metric
	PopAllVmMetrics(ctx context.Context, ref objects.ManagedObjectReference) []*objects.Metric
}
