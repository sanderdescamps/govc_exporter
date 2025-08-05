package collector

import (
	"context"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

type VMPerfCollector struct {
	extraLabels []string
	scraper     *scraper.VCenterScraper

	perfMetric *prometheus.Desc
}

func NewVMPerfCollector(scraper *scraper.VCenterScraper, cConf Config) *VMPerfCollector {
	labels := []string{"uuid", "name", "template", "vm_id"}
	extraLabels := cConf.VMTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	perfLabels := append(labels, "kind", "unit")

	return &VMPerfCollector{
		scraper:     scraper,
		extraLabels: extraLabels,
		perfMetric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "perf_metric"),
			"Performance metric", perfLabels, nil),
	}
}

func (c *VMPerfCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.perfMetric
}

func (c *VMPerfCollector) Collect(ch chan<- prometheus.Metric) {
	if c.scraper.Host == nil || c.scraper.VM == nil || c.scraper.VMPerf == nil {
		return
	}
	ctx := context.Background()
	for vm := range c.scraper.DB.GetAllVMIter(ctx) {
		extraLabelValues := []string{}
		objectTags := c.scraper.DB.GetTags(ctx, vm.Self)
		for _, tagCat := range c.extraLabels {
			extraLabelValues = append(extraLabelValues, objectTags.GetTag(tagCat))
		}

		labelValues := []string{vm.UUID, vm.Name, strconv.FormatBool(vm.Template), vm.Self.ID()}
		labelValues = append(labelValues, extraLabelValues...)

		for _, metric := range c.scraper.MetricsDB.PopAllVmMetrics(ctx, vm.Self) {
			perfMetricLabelValues := append(labelValues, metric.Name, metric.Unit)
			ch <- prometheus.NewMetricWithTimestamp(metric.Timestamp, prometheus.MustNewConstMetric(
				c.perfMetric, prometheus.GaugeValue, metric.Value, perfMetricLabelValues...,
			))
		}
	}
}
