package scraper

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	metricshelper "github.com/sanderdescamps/govc_exporter/internal/scraper/metrics_helper"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/exp/constraints"
)

type perfQuery struct {
	metrics        []string
	instance       string
	maxSample      int32
	sampleInterval time.Duration
	// window         time.Duration
	begin time.Time
	end   time.Time
}

func (pq perfQuery) ToSpec() types.PerfQuerySpec {
	return types.PerfQuerySpec{
		MaxSample:  pq.maxSample,
		MetricId:   []types.PerfMetricId{{Instance: pq.instance}},
		IntervalId: int32(pq.sampleInterval.Seconds()),
		StartTime:  &pq.begin,
		EndTime:    &pq.end,
	}
}

type PerfOption func(*perfQuery)

func NewPerfQuery(options ...PerfOption) *perfQuery {
	result := &perfQuery{
		metrics:   []string{},
		instance:  "*",
		maxSample: 1,
		// sampleInterval: 20,
	}
	for _, o := range options {
		o(result)
	}
	return result
}

func SetMaxSamples[T constraints.Integer](num T) PerfOption {
	return func(pq *perfQuery) {
		pq.maxSample = int32(num)
	}
}

func SetInterval(d time.Duration) PerfOption {
	return func(pq *perfQuery) {
		pq.sampleInterval = d
	}
}

func SetDurationWindow(d time.Duration) PerfOption {
	endtime := time.Now()
	starttime := endtime.Add(d)
	return func(pq *perfQuery) {
		pq.begin = starttime
		pq.end = endtime
	}
}

func SetWindow(t1 time.Time, t2 time.Time) PerfOption {
	if t1.After(t2) {
		t1, t2 = t2, t1
	}

	return func(pq *perfQuery) {
		pq.begin = t1
		pq.end = t2
	}
}

func AddMetrics(m ...string) PerfOption {
	return func(pq *perfQuery) {
		pq.metrics = append(pq.metrics, m...)
	}
}

func SetMetrics(m ...string) PerfOption {
	return func(pq *perfQuery) {
		pq.metrics = m
	}
}

func SetInstance(i string) PerfOption {
	return func(pq *perfQuery) {
		pq.instance = i
	}
}

func EntityMetricToMetric(entiry performance.EntityMetric) []objects.Metric {
	result := []objects.Metric{}
	for id, sample := range entiry.SampleInfo {
		for _, entityValue := range entiry.Value {
			result = append(result,
				objects.Metric{
					Ref:      objects.NewManagedObjectReferenceFromVMwareRef(entiry.Entity),
					Name:     entityValue.Name,
					Unit:     entityValue.Unit,
					Instance: entityValue.Instance,
					Value: func() float64 {
						if entityValue.Value != nil {
							if len(entiry.SampleInfo) == len(entityValue.Value) {
								return float64(entityValue.Value[id])
							} else {
								return Avg(entityValue.Value)
							}
						}
						return 0
					}(),
					Timestamp: sample.Timestamp,
				},
			)
		}

	}
	return result
}

// BasePerfSensor

type BasePerfSensor struct {
	// perfMetrics   map[types.ManagedObjectReference]*MetricQueue
	// scraper       *VCenterScraper
	lastQueryTime time.Time
	// sensorKind    string
	sensorLock sync.Mutex
	config     PerfSensorConfig
	metrics    []string
}

func NewBasePerfSensor(config PerfSensorConfig, metrics []string) *BasePerfSensor {
	return &BasePerfSensor{
		config:  config,
		metrics: metrics,
	}
}

func (s *BasePerfSensor) QueryEntiryMetrics(ctx context.Context, scraper *VCenterScraper, refs []types.ManagedObjectReference) ([]performance.EntityMetric, metricshelper.RefreshStats, error) {
	if ok := s.sensorLock.TryLock(); !ok {
		return nil, metricshelper.RefreshStats{
			Failed: true,
		}, fmt.Errorf("Sensor already running")
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
		SetMetrics(s.metrics...),
	)

	pq := NewPerfQuery(options...)
	t1 := time.Now()
	client, release, err := scraper.clientPool.Acquire()
	defer release()
	if err != nil {
		return nil, metricshelper.RefreshStats{
			Failed: true,
		}, err
	}
	t2 := time.Now()
	perfManager := performance.NewManager(client.Client)

	sample, err := perfManager.SampleByName(ctx, pq.ToSpec(), pq.metrics, refs)
	t3 := time.Now()
	refreshStats := metricshelper.RefreshStats{
		ClientWaitTime: t2.Sub(t1),
		QueryTime:      t3.Sub(t2),
		Failed:         false,
	}

	if err != nil {
		refreshStats.Failed = true
		return nil, refreshStats, err
	}

	metricSeries, err := perfManager.ToMetricSeries(ctx, sample)
	if err != nil {
		refreshStats.Failed = true
		return nil, refreshStats, err
	}
	s.lastQueryTime = windowEnd

	return metricSeries, refreshStats, nil
}
