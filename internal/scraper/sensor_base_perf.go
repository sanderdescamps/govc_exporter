package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/types"
)

type BasePerfSensor struct {
	perfMetrics   map[types.ManagedObjectReference]*MetricQueue
	scraper       *VCenterScraper
	lastQueryTime time.Time
	sensorKind    string
	sensorLock    sync.Mutex
	config        PerfSensorConfig
	metrics       struct {
		QueryTime      *SensorMetricDuration
		ClientWaitTime *SensorMetricDuration
		Status         *SensorMetricStatus
	}
}

func NewBasePerfSensor(scraper *VCenterScraper, sensorKind string, config PerfSensorConfig) *BasePerfSensor {
	sensor := BasePerfSensor{
		sensorKind:    sensorKind,
		perfMetrics:   map[types.ManagedObjectReference]*MetricQueue{},
		scraper:       scraper,
		config:        config,
		lastQueryTime: time.Now().Add(-5 * time.Minute),
		metrics: struct {
			QueryTime      *SensorMetricDuration
			ClientWaitTime *SensorMetricDuration
			Status         *SensorMetricStatus
		}{
			QueryTime:      NewSensorMetricDuration(sensorKind, "client_wait_time", 0),
			ClientWaitTime: NewSensorMetricDuration(sensorKind, "query_time", 0),
			Status:         NewSensorMetricStatus(sensorKind, "status", false),
		},
	}
	scraper.RegisterSensorMetric(
		&sensor.metrics.ClientWaitTime.SensorMetric,
		&sensor.metrics.QueryTime.SensorMetric,
		&sensor.metrics.Status.SensorMetric,
	)
	return &sensor
}

func (s *BasePerfSensor) Clean(maxAge time.Duration, logger *slog.Logger) {
	now := time.Now()
	total := 0
	for _, pm := range s.perfMetrics {
		ExpiredMetrics := pm.PopOlderOrEqualThan(now.Add(-maxAge))
		total = total + len(ExpiredMetrics)
	}
	if total > 0 {
		logger.Warn(fmt.Sprintf("Removed %d host-metrics which were not yet pulled", total))
	}
}

func (s *BasePerfSensor) PopAll(ref types.ManagedObjectReference) []*Metric {
	if perfMetrics, ok := s.perfMetrics[ref]; ok {
		return perfMetrics.PopAll()
	}
	return nil
}

func (s *BasePerfSensor) GetAllJsons() (map[string][]byte, error) {
	result := map[string][]byte{}

	for ref, queue := range s.perfMetrics {
		name := ref.Value
		jsonBytes, err := json.MarshalIndent(queue, "", "  ")
		if err != nil {
			return nil, err
		}
		result[name] = jsonBytes
	}
	return result, nil
}

func (s *BasePerfSensor) QueryEntiryMetrics(refs []types.ManagedObjectReference, metrics []string, ctx context.Context, logger *slog.Logger) ([]performance.EntityMetric, error) {
	if ok := s.sensorLock.TryLock(); !ok {
		return nil, fmt.Errorf("Sensor already running")
	}
	defer s.sensorLock.Unlock()
	windowEnd := time.Now()
	windowBegin := windowEnd.Add(-s.config.MaxSampleWindow)
	if s.lastQueryTime.After(windowBegin) {
		windowBegin = s.lastQueryTime
	}
	options := []PerfOption{}
	options = append(options,
		SetMaxSamples(20),
		SetInterval(s.config.SampleInterval),
		SetWindow(windowBegin, windowEnd),
		SetMetrics(metrics...),
	)

	pq := NewPerfQuery(options...)
	t1 := time.Now()
	client, release, err := s.scraper.clientPool.Acquire()
	defer release()
	if err != nil {
		return nil, err
	}
	t2 := time.Now()
	perfManager := performance.NewManager(client.Client)

	sample, err := perfManager.SampleByName(ctx, pq.ToSpec(), pq.metrics, refs)
	t3 := time.Now()
	s.metrics.ClientWaitTime.Update(t2.Sub(t1))
	s.metrics.QueryTime.Update(t3.Sub(t2))
	s.metrics.Status.Update(true)
	if err != nil {
		s.metrics.Status.Update(true)
		return nil, err
	}

	metricSeries, err := perfManager.ToMetricSeries(ctx, sample)
	if err != nil {
		return nil, err
	}
	s.lastQueryTime = windowEnd

	return metricSeries, nil
}

func (s *BasePerfSensor) Kind() string {
	return s.sensorKind
}
