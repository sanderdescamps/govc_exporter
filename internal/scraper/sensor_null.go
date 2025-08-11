package scraper

import (
	"context"

	sensormetrics "github.com/sanderdescamps/govc_exporter/internal/scraper/sensor_metrics"
)

type NullSensor struct {
	kind string
}

func NewNullSensor(kind string) *NullSensor {
	return &NullSensor{
		kind: kind,
	}
}

func (s *NullSensor) Init(ctx context.Context, scraper *VCenterScraper) error {
	return nil
}

func (s *NullSensor) StartRefresher(ctx context.Context, scraper *VCenterScraper) error {
	return nil
}

func (s *NullSensor) StopRefresher(ctx context.Context) {
}

func (s *NullSensor) Enabled() bool {
	return false
}

func (s *NullSensor) Kind() string {
	return s.kind
}

func (s *NullSensor) GetLatestMetrics() []sensormetrics.SensorMetric {
	return []sensormetrics.SensorMetric{
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
}
