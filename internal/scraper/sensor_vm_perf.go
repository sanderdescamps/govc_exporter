package scraper

import (
	"context"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
)

type VMPerfSensor struct {
	BasePerfSensor
	AutoRunSensor
	Refreshable
	helper.Matchable
}

func NewVMPerfSensor(scraper *VCenterScraper, config PerfSensorConfig) *VMPerfSensor {
	var sensor VMPerfSensor
	sensor = VMPerfSensor{
		BasePerfSensor: *NewBasePerfSensor(scraper, "VMPerfSensor", config),
		AutoRunSensor:  *NewAutoRunSensor(&sensor, config.SensorConfig()),
	}
	return &sensor
}

func (s *VMPerfSensor) Refresh(ctx context.Context) error {
	s.scraper.VM.WaitTillStartup()
	refs := s.scraper.VM.GetAllRefs()
	if len(refs) < 1 {
		return nil
	}

	metrics := []string{}
	if s.config.DefaultMetrics {
		metrics = append(metrics,
			"cpu.usagemhz.average",
			"cpu.capacity.provisioned.average",
			"cpu.readiness.average",
			"cpu.costop.summation",
			"cpu.maxlimited.summation",
			"cpu.ready.summation",
			"mem.active.average",
			"mem.granted.average",
			"mem.consumed.average",
			"disk.throughput.contention.average",
			"disk.throughput.usage.average",
		)
	}
	if len(s.config.ExtraMetrics) > 0 {
		metrics = append(metrics, s.config.ExtraMetrics...)
	}

	metrics = helper.Dedup(metrics)

	metricSeries, err := s.BasePerfSensor.QueryEntiryMetrics(ctx, refs, metrics)
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
	return nil
}

func (s *VMPerfSensor) Name() string {
	return "perf-vm"
}

func (s *VMPerfSensor) Match(name string) bool {
	return helper.NewMatcher("perf-vm", "perfvm", "vm-perf", "vmperf").Match(name)
}
