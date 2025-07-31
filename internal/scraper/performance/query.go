package perfmetrics

import (
	"time"

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
