package scraper

import (
	"context"
	"log/slog"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/types"
)

type HostPerfSensor struct {
	BasePerfSensor
	Refreshable
	helper.Matchable
}

func NewHostPerfSensor(scraper *VCenterScraper, options ...PerfOption) *HostPerfSensor {
	sensor := &HostPerfSensor{
		BasePerfSensor: BasePerfSensor{
			perfMetrics: map[types.ManagedObjectReference]*MetricQueue{},
			scraper:     scraper,
			perfOptions: options,
			metrics: struct {
				QueryTime      *SensorMetricDuration
				ClientWaitTime *SensorMetricDuration
				Status         *SensorMetricStatus
			}{
				QueryTime:      NewSensorMetricDuration("sensor.host_perf.client_wait_time", 0),
				ClientWaitTime: NewSensorMetricDuration("sensor.host_perf.query_time", 0),
				Status:         NewSensorMetricStatus("sensor.host_perf.status", false),
			},
		},
	}
	scraper.RegisterSensorMetric(
		&sensor.metrics.ClientWaitTime.SensorMetric,
		&sensor.metrics.QueryTime.SensorMetric,
		&sensor.metrics.Status.SensorMetric,
	)
	return sensor
}

func (s *HostPerfSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
	t_start := time.Now()
	options := []PerfOption{}
	options = append(options,
		SetSamples(20),
		// SetInterval(60*time.Second),
		SetWindow(5*time.Minute),
		SetMetrics(
			"cpu.usagemhz.average",
			"cpu.demand.average",
			"cpu.latency.average",
			"cpu.entitlement.latest",
			"cpu.ready.summation",
			"cpu.readiness.average",
			"cpu.costop.summation",
			"cpu.maxlimited.summation",
			"mem.entitlement.average",
			"mem.active.average",
			"mem.shared.average",
			"mem.vmmemctl.average",
			"mem.swapped.average",
			"mem.consumed.average",
			"net.bytesRx.average",
			"net.bytesTx.average",
			"net.errorsRx.summation",
			"net.errorsTx.summation",
			"net.droppedRx.summation",
			"net.droppedTx.summation",
			"datastore.read.average",
			"datastore.write.average",
			// "datastore.numberReadAveraged.average",
			// "datastore.numberWriteAveraged.average",
			// "datastore.totalReadLatency.average",
			// "datastore.totalWriteLatency.average",
		),
	)

	options = append(options, s.perfOptions...)

	pq := NewPerfQuery(options...)
	t1 := time.Now()
	client, release, err := s.scraper.clientPool.Acquire()
	defer release()
	if err != nil {
		return err
	}
	t2 := time.Now()
	perfManager := performance.NewManager(client.Client)
	hostRefs := s.scraper.Host.GetAllRefs()
	if len(hostRefs) == 0 {
		return nil
	}
	sample, err := perfManager.SampleByName(ctx, pq.ToSpec(), pq.metrics, hostRefs)
	t3 := time.Now()
	s.metrics.ClientWaitTime.Update(t2.Sub(t1))
	s.metrics.QueryTime.Update(t3.Sub(t2))
	s.metrics.Status.Update(true)
	if err != nil {
		s.metrics.Status.Update(true)
		return err
	}

	metricSeries, err := perfManager.ToMetricSeries(ctx, sample)
	if err != nil {
		return err
	}

	for _, metricSerie := range metricSeries {
		entity := metricSerie.Entity
		if _, ok := s.perfMetrics[entity]; !ok {
			s.perfMetrics[entity] = NewMetricQueue()
		}
		for _, metric := range EntityMetricToMetric(metricSerie) {
			s.perfMetrics[entity].Add(metric)
		}
	}
	t_end := time.Now()
	logger.Debug("host performance metrics collected", "time_ms", t_end.Sub(t_start).Milliseconds())
	return nil
}

func (s *HostPerfSensor) Name() string {
	return "perf-host"
}

func (s *HostPerfSensor) Match(name string) bool {
	return helper.NewMatcher("perf-host", "perfhost", "perfesx", "perf-esx", "host-perf-host", "hostperf", "esxperf", "esx-perf").Match(name)
}
