package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

type esxPerfCollector struct {
	extraLabels []string
	scraper     *scraper.VCenterScraper

	perfMetric *prometheus.Desc
}

func NewEsxPerfCollector(scraper *scraper.VCenterScraper, cConf Config) *esxPerfCollector {
	labels := []string{"id", "name", "datacenter", "cluster"}
	extraLabels := cConf.HostTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	perfLabels := append(labels, "kind", "unit")

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
	if c.scraper.Host == nil || c.scraper.HostPerf == nil {
		return
	}

	for _, ref := range c.scraper.Host.GetAllRefs() {
		host := c.scraper.Host.Get(ref)
		if host == nil {
			continue
		}

		parentChain := c.scraper.GetParentChain(host.Self)

		extraLabelValues := func() []string {
			result := []string{}
			for _, tagCat := range c.extraLabels {
				tag := c.scraper.Tags.GetTag(host.Self, tagCat)
				if tag != nil {
					result = append(result, tag.Name)
				} else {
					result = append(result, "")
				}
			}
			return result
		}()
		labelValues := []string{me2id(host.ManagedEntity), host.Name, parentChain.DC, parentChain.Cluster}
		labelValues = append(labelValues, extraLabelValues...)

		for metric := range c.scraper.HostPerf.PopAllItems(ref) {
			perfMetricLabelValues := append(labelValues, metric.Name, metric.Unit)
			ch <- prometheus.NewMetricWithTimestamp(metric.TimeStamp, prometheus.MustNewConstMetric(
				c.perfMetric, prometheus.GaugeValue, metric.Value, perfMetricLabelValues...,
			))
		}
	}
}
