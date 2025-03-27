package scraper

import (
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/timequeue"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/exp/constraints"
)

type Metric struct {
	timequeue.Event
	Ref       types.ManagedObjectReference `json:"ref"`
	Name      string                       `json:"name"`
	Unit      string                       `json:"unit"`
	Instance  string                       `json:"instance"`
	Value     float64                      `json:"value"`
	TimeStamp time.Time                    `json:"time_stamp"`
}

func (m Metric) GetTimestamp() time.Time {
	return m.TimeStamp
}

func EntityMetricToMetric(entiry performance.EntityMetric) []Metric {
	result := []Metric{}
	for id, sample := range entiry.SampleInfo {
		for _, entityValue := range entiry.Value {
			result = append(result,
				Metric{
					Ref:      entiry.Entity,
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
					TimeStamp: sample.Timestamp,
				},
			)
		}

	}
	return result
}

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
