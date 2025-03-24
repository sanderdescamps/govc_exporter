package scraper

import (
	"context"
	"log/slog"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/types"
)

type VMPerfSensor struct {
	BasePerfSensor
	Refreshable
	helper.Matchable
}

func NewVMPerfSensor(scraper *VCenterScraper) *VMPerfSensor {
	sensor := &VMPerfSensor{
		BasePerfSensor: BasePerfSensor{
			perfMetrics: map[types.ManagedObjectReference]*MetricQueue{},
			scraper:     scraper,
			metrics: struct {
				QueryTime      *SensorMetricDuration
				ClientWaitTime *SensorMetricDuration
				Status         *SensorMetricStatus
			}{
				QueryTime:      NewSensorMetricDuration("sensor.vm_perf.client_wait_time", 0),
				ClientWaitTime: NewSensorMetricDuration("sensor.vm_perf.query_time", 0),
				Status:         NewSensorMetricStatus("sensor.vm_perf.status", false),
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

func (s *VMPerfSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
	t_start := time.Now()
	options := []PerfOption{}
	options = append(options,
		SetSamples(20),
		SetWindow(5*time.Minute),
		SetMetrics(
			"cpu.usagemhz.average",
			"cpu.capacity.provisioned.average",
			"cpu.readiness.average",
			"cpu.costop.summation",
			"mem.active.average",
			"mem.granted.average",
			"mem.consumed.average",
			"disk.throughput.contention.average",
			"disk.throughput.usage.average",
		),
	)

	pq := NewPerfQuery(options...)
	t1 := time.Now()
	client, release, err := s.scraper.clientPool.Acquire()
	defer release()
	if err != nil {
		return err
	}
	t2 := time.Now()
	perfManager := performance.NewManager(client.Client)
	vmRefs := s.scraper.VM.GetAllRefs()
	if len(vmRefs) == 0 {
		return nil
	}
	sample, err := perfManager.SampleByName(ctx, pq.ToSpec(), pq.metrics, vmRefs)
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
	logger.Debug("vm performance metrics collected", "time_ms", t_end.Sub(t_start).Milliseconds())
	return nil
}

func (s *VMPerfSensor) Name() string {
	return "perf-vm"
}

func (s *VMPerfSensor) Match(name string) bool {
	return helper.NewMatcher("perf-vm", "perfvm", "vm-perf", "vmperf").Match(name)
}
