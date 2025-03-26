package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

const (
	generalCollectorSubsystem = "scraper"
)

type scraperCollector struct {
	scraper       *scraper.VCenterScraper
	scraperMetric *prometheus.Desc
}

func NewScraperCollector(scraper *scraper.VCenterScraper) *scraperCollector {
	sensorLabels := []string{"sensor", "metric", "unit"}
	return &scraperCollector{
		scraper: scraper,
		scraperMetric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, generalCollectorSubsystem, "metric"),
			"Metric about the scraper inside the exporter", sensorLabels, nil),
	}
}

func (c *scraperCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.scraperMetric
}

func (c *scraperCollector) Collect(ch chan<- prometheus.Metric) {

	for _, metric := range c.scraper.ScraperMetrics() {
		labelValues := []string{metric.Sensor, metric.MetricName, metric.Unit}
		ch <- prometheus.MustNewConstMetric(
			c.scraperMetric, prometheus.GaugeValue, metric.Value, labelValues...,
		)
	}
}
