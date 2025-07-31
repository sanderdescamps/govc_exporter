package scraper

import (
	"context"

	metricshelper "github.com/sanderdescamps/govc_exporter/internal/scraper/metrics_helper"
)

type NullSensor struct {
	kind string
}

func NewNullSensor(kind string) *NullSensor {
	return &NullSensor{
		kind: kind,
	}
}

func (s *NullSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) {
	return
}

func (s *NullSensor) StopRefresher(ctx context.Context) {
	return
}

func (s *NullSensor) Enabled() bool {
	return false
}

func (s *NullSensor) Kind() string {
	return s.kind
}

func (s *NullSensor) GetLatestMetrics() []metricshelper.SensorMetric {
	return []metricshelper.SensorMetric{
		{
			Sensor:     s.kind,
			MetricName: "enabled",
			Value:      0.0,
			Unit:       "boolean",
		},
	}
}

// Match implements Sensor.
func (s *NullSensor) Match(string) bool {
	return false
}

// TriggerManualRefresh implements Sensor.
func (s *NullSensor) TriggerManualRefresh(ctx context.Context) {
	return
}
