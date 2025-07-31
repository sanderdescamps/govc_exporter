package scraper

import (
	"context"

	metricshelper "github.com/sanderdescamps/govc_exporter/internal/scraper/metrics_helper"
)

type Sensor interface {
	StartRefresher(ctx context.Context, scraper *VCenterScraper)
	StopRefresher(ctx context.Context)
	Enabled() bool
	Kind() string
	Match(string) bool
	TriggerManualRefresh(ctx context.Context)
	GetLatestMetrics() []metricshelper.SensorMetric
}
