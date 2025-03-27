package scraper

import (
	"context"
	"log/slog"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
)

type HostPerfSensor struct {
	BasePerfSensor
	AutoRunSensor
	Refreshable
	helper.Matchable
}

func NewHostPerfSensor(scraper *VCenterScraper, config PerfSensorConfig) *HostPerfSensor {
	var sensor HostPerfSensor
	sensor = HostPerfSensor{
		BasePerfSensor: *NewBasePerfSensor(scraper, "HostPerfSensor", config),
		AutoRunSensor:  *NewAutoRunSensor(&sensor, config.SensorConfig()),
	}
	return &sensor
}

func (s *HostPerfSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
	s.scraper.Host.WaitTillStartup()
	hostRefs := s.scraper.Host.GetAllRefs()
	if len(hostRefs) < 1 {
		return nil
	}

	metrics := []string{}
	if s.config.DefaultMetrics {
		metrics = append(metrics,
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
		)
	}
	if len(s.config.ExtraMetrics) > 0 {
		metrics = append(metrics, s.config.ExtraMetrics...)
	}

	metrics = helper.Dedup(metrics)

	metricSeries, err := s.BasePerfSensor.QueryEntiryMetrics(hostRefs, metrics, ctx, logger)
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

func (s *HostPerfSensor) Name() string {
	return "perf-host"
}

func (s *HostPerfSensor) Match(name string) bool {
	return helper.NewMatcher("perf-host", "perfhost", "perfesx", "perf-esx", "host-perf-host", "hostperf", "esxperf", "esx-perf").Match(name)
}
