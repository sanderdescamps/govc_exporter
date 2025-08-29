package scraper

import (
	"context"
	"fmt"
	"iter"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	sensormetrics "github.com/sanderdescamps/govc_exporter/internal/scraper/sensor_metrics"
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
	config     config.PerfSensorConfig
	metrics    []string

	metricsCollector *sensormetrics.SensorMetricsCollector
	statusMonitor    *sensormetrics.StatusMonitor
}

func NewBasePerfSensor(config config.PerfSensorConfig, metrics []string, mc *sensormetrics.SensorMetricsCollector, sm *sensormetrics.StatusMonitor) *BasePerfSensor {
	return &BasePerfSensor{
		config:           config,
		metrics:          metrics,
		metricsCollector: mc,
		statusMonitor:    sm,
	}
}

func (s *BasePerfSensor) QueryVMwareEntiryMetrics(ctx context.Context, scraper *VCenterScraper, refs []types.ManagedObjectReference) ([]performance.EntityMetric, error) {
	if ok := s.sensorLock.TryLock(); !ok {
		return nil, fmt.Errorf("Sensor already running")
	}
	defer s.sensorLock.Unlock()

	sensorStopwatch := sensormetrics.NewSensorStopwatch()

	windowEnd := time.Now().Truncate(s.config.SampleInterval)
	windowBegin1 := s.lastQueryTime.Add(s.config.SampleInterval)
	windowBegin2 := windowEnd.Add(-s.config.MaxSampleWindow)

	var windowBegin time.Time = windowBegin2
	if windowBegin1.After(windowBegin2) {
		windowBegin = windowBegin1
	}
	options := []PerfOption{}
	options = append(options,
		SetMaxSamples(20),
		SetInterval(s.config.SampleInterval),
		SetWindow(windowBegin, windowEnd),
		SetMetrics(s.metrics...),
	)

	pq := NewPerfQuery(options...)
	sensorStopwatch.Start()
	client, release, err := scraper.clientPool.Acquire()
	defer release()
	if err != nil {
		return nil, err
	}
	sensorStopwatch.Mark1()
	perfManager := performance.NewManager(client.Client)

	sample, err := perfManager.SampleByName(ctx, pq.ToSpec(), pq.metrics, refs)
	sensorStopwatch.Finish()
	s.metricsCollector.UploadStats(sensorStopwatch.GetStats())

	if err != nil {
		return nil, err
	}

	metricSeries, err := perfManager.ToMetricSeries(ctx, sample)
	if err != nil {
		return nil, err
	}
	s.lastQueryTime = windowEnd

	return metricSeries, nil
}

func (s *BasePerfSensor) ToFilteredMetrics(entityMetrics []performance.EntityMetric, filterConf []config.PerfFilter) (ret map[objects.ManagedObjectReference][]objects.Metric) {
	filter := PerfMetricFilterFromConf(filterConf)
	ret = make(map[objects.ManagedObjectReference][]objects.Metric)
	for _, metricSerie := range entityMetrics {
		entityRef := objects.NewManagedObjectReferenceFromVMwareRef(metricSerie.Entity)
		metrics := EntityMetricToMetric(metricSerie)
		metrics = filter(metrics)
		ret[entityRef] = append(ret[entityRef], metrics...)
	}
	return
}

func (s *BasePerfSensor) ToFilteredMetricsIter(entityMetrics []performance.EntityMetric, filterConf []config.PerfFilter) iter.Seq2[objects.ManagedObjectReference, []objects.Metric] {
	filter := PerfMetricFilterFromConf(filterConf)
	return func(yield func(objects.ManagedObjectReference, []objects.Metric) bool) {
		for _, metricSerie := range entityMetrics {
			entityRef := objects.NewManagedObjectReferenceFromVMwareRef(metricSerie.Entity)
			metrics := EntityMetricToMetric(metricSerie)
			metrics = filter(metrics)
			if !yield(entityRef, metrics) {
				return
			}
		}
	}
}

func PerfMetricFilterFromConf(filterConf []config.PerfFilter) func(m []objects.Metric) []objects.Metric {
	filters := []func(m []objects.Metric) []objects.Metric{}
	for _, fConf := range filterConf {
		switch fConf.Action {
		case config.PerfFilterActionDrop:
			filters = append(filters, PerfMetricDropFilter(fConf.MatchName, fConf.MatchInstance))
		case config.PerfFilterActionSum:
			filters = append(filters, PerfMetricSumFilter(fConf.MatchName, fConf.MatchInstance, fConf.NewName))
		case config.PerfFilterActionSpit:
			filters = append(filters, PerfMetricSplitToMultiInstanceFilter(fConf.MatchName, fConf.MatchInstance, strings.Split(fConf.NewName, ",")))
		case config.PerfFilterActionRename:
			filters = append(filters, PerfMetricRenameFilter(fConf.MatchName, fConf.MatchInstance, fConf.NewName))
		case config.PerfFilterActionInstanceRename:
			filters = append(filters, PerfMetricInstanceRenameFilter(fConf.MatchName, fConf.MatchInstance, fConf.NewName))
		}
	}

	return func(m []objects.Metric) []objects.Metric {
		result := m
		for _, f := range filters {
			result = f(result)
		}
		return result
	}
}

// Filter only intended for testing
// Duplicates metrics with diffrent instance names.
func PerfMetricSplitToMultiInstanceFilter(matchName regexp.Regexp, matchInstance regexp.Regexp, instances []string) func([]objects.Metric) []objects.Metric {
	return func(metrics []objects.Metric) (ret []objects.Metric) {
		for _, metric := range metrics {
			if matchName.MatchString(metric.Name) && matchInstance.MatchString(metric.Instance) {
				for _, i := range instances {
					ret = append(ret, objects.Metric{
						Ref:       metric.Ref,
						Name:      metric.Name,
						Unit:      metric.Unit,
						Instance:  i,
						Value:     metric.Value,
						Timestamp: metric.Timestamp,
					})
				}
			} else {
				ret = append(ret, metric)
			}
		}
		return
	}
}

// Filter only intended for testing
func PerfMetricRenameFilter(matchName regexp.Regexp, matchInstance regexp.Regexp, new_name string) func([]objects.Metric) []objects.Metric {
	return func(metrics []objects.Metric) (ret []objects.Metric) {
		for _, metric := range metrics {
			if matchName.MatchString(metric.Name) && matchInstance.MatchString(metric.Instance) {
				metric.Name = new_name
			}
			ret = append(ret, metric)
		}
		return
	}
}

// Filter only intended for testing
func PerfMetricInstanceRenameFilter(matchName regexp.Regexp, matchInstance regexp.Regexp, new_name string) func([]objects.Metric) []objects.Metric {
	return func(metrics []objects.Metric) (ret []objects.Metric) {
		for _, metric := range metrics {
			if matchName.MatchString(metric.Name) && matchInstance.MatchString(metric.Instance) {
				metric.Instance = new_name
			}
			ret = append(ret, metric)
		}
		return
	}
}

func PerfMetricDropFilter(matchName regexp.Regexp, matchInstance regexp.Regexp) func([]objects.Metric) []objects.Metric {
	return func(metrics []objects.Metric) (ret []objects.Metric) {
		for _, metric := range metrics {
			if !matchName.MatchString(metric.Name) || !matchInstance.MatchString(metric.Instance) {
				ret = append(ret, metric)
			}
		}
		return
	}
}

func PerfMetricSumFilter(matchName regexp.Regexp, matchInstance regexp.Regexp, newName string) func([]objects.Metric) []objects.Metric {
	newName = strings.TrimSuffix(newName, ".sum")

	return func(metrics []objects.Metric) (ret []objects.Metric) {
		sumMap := make(map[time.Time]*objects.Metric)
		for _, metric := range metrics {
			if matchName.MatchString(metric.Name) && matchInstance.MatchString(metric.Instance) {
				if sumMetric, ok := sumMap[metric.Timestamp]; !ok {
					sumMap[metric.Timestamp] = &objects.Metric{
						Name:      metric.Name,
						Ref:       metric.Ref,
						Unit:      metric.Unit,
						Timestamp: metric.Timestamp,
						Value:     metric.Value,
						Instance:  metric.Instance,
					}
				} else if sumMetric.Ref != metric.Ref {
					panic("all metrics must be from the same reference")
				} else if sumMetric.Unit == metric.Unit {
					sumMap[metric.Timestamp].Value += metric.Value
					if sumMetric.Instance != metric.Instance {
						sumMetric.Instance = ""
					}
					if sumMetric.Name != metric.Name {
						if newName != "" {
							sumMetric.Name = newName
						} else if strings.HasPrefix(sumMetric.Name, "summation.metric.") {
							// keep sumMetric.Name
						} else if commonPrefix := helper.CommonPrefix(sumMetric.Name, metric.Name); commonPrefix != "" {
							sumMetric.Name = commonPrefix
						} else {
							sumMetric.Name = fmt.Sprintf("summation.metric.%s", helper.HashStrings(matchName.String(), matchInstance.String()))
						}
					}
				}
			} else {
				ret = append(ret, metric)
			}
		}

		for _, metric := range sumMap {
			metric.Name = metric.Name + ".sum"
			ret = append(ret, *metric)
		}
		return
	}
}
