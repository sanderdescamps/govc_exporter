package scraper

import (
	"context"

	sensormetrics "github.com/sanderdescamps/govc_exporter/internal/scraper/sensor_metrics"
)

type Sensor interface {
	Init(ctx context.Context, scraper *VCenterScraper) error
	StartRefresher(ctx context.Context, scraper *VCenterScraper) error
	StopRefresher(ctx context.Context)
	Enabled() bool
	Kind() string
	Match(string) bool
	TriggerManualRefresh(ctx context.Context)
	GetLatestMetrics() []sensormetrics.SensorMetric
}
