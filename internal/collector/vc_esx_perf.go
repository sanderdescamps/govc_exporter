package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

type esxPerfCollector struct {
	extraLabels []string
	scraper     *scraper.VCenterScraper

	perfMetric *prometheus.Desc
}

func NewEsxPerfCollector(scraper *scraper.VCenterScraper, cConf config.CollectorConfig) *esxPerfCollector {
	labels := []string{"id", "name", "datacenter", "cluster"}
	extraLabels := cConf.HostTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	perfLabels := append(labels, "kind", "instance", "unit")

	return &esxPerfCollector{
		scraper:     scraper,
		extraLabels: extraLabels,
		perfMetric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "perf_metric"),
			"Performance metric", perfLabels, nil),
	}
}

func (c *esxPerfCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.perfMetric
}

func (c *esxPerfCollector) Collect(ch chan<- prometheus.Metric) {
	if !c.scraper.Host.Enabled() || !c.scraper.HostPerf.Enabled() {
		return
	}
	ctx := context.Background()
	for host := range c.scraper.DB.GetAllHostIter(ctx) {

		extraLabelValues := []string{}
		objectTags := c.scraper.DB.GetTags(ctx, host.Self)
		for _, tagCat := range c.extraLabels {
			extraLabelValues = append(extraLabelValues, objectTags.GetTag(tagCat))
		}

		labelValues := []string{host.Self.ID(), host.Name, host.Datacenter, host.Cluster}
		labelValues = append(labelValues, extraLabelValues...)

		for metric := range c.scraper.MetricsDB.PopAllHostMetricsIter(ctx, host.Self) {
			perfMetricLabelValues := append(labelValues, metric.Name, metric.Instance, metric.Unit)
			ch <- prometheus.NewMetricWithTimestamp(metric.Timestamp, prometheus.MustNewConstMetric(
				c.perfMetric, prometheus.GaugeValue, metric.Value, perfMetricLabelValues...,
			))
		}
	}
}
