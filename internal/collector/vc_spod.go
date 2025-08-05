package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

const (
	storagePodCollectorSubsystem = "spod"
)

type storagePodCollector struct {
	scraper     *scraper.VCenterScraper
	extraLabels []string
	capacity    *prometheus.Desc
	freeSpace   *prometheus.Desc
}

func NewStoragePodCollector(scraper *scraper.VCenterScraper, cConf Config) *storagePodCollector {
	labels := []string{"id", "name", "datacenter"}

	extraLabels := cConf.StoragePodTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	return &storagePodCollector{
		scraper:     scraper,
		extraLabels: extraLabels,
		capacity: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, storagePodCollectorSubsystem, "capacity_bytes"),
			"storagePod capacity in bytes", labels, nil),
		freeSpace: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, storagePodCollectorSubsystem, "free_space_bytes"),
			"storagePod freespace in bytes", labels, nil),
	}
}

func (c *storagePodCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.capacity
	ch <- c.freeSpace
}

func (c *storagePodCollector) Collect(ch chan<- prometheus.Metric) {
	if !c.scraper.SPOD.Enabled() {
		return
	}
	ctx := context.Background()
	for spod := range c.scraper.DB.GetAllStoragePodIter(ctx) {

		extraLabelValues := []string{}
		objectTags := c.scraper.DB.GetTags(ctx, spod.Self)
		for _, tagCat := range c.extraLabels {
			extraLabelValues = append(extraLabelValues, objectTags.GetTag(tagCat))
		}

		labelValues := []string{spod.Self.ID(), spod.Name, spod.Datacenter}
		labelValues = append(labelValues, extraLabelValues...)
		ch <- prometheus.NewMetricWithTimestamp(spod.Timestamp, prometheus.MustNewConstMetric(
			c.capacity, prometheus.GaugeValue, float64(spod.Capacity), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(spod.Timestamp, prometheus.MustNewConstMetric(
			c.freeSpace, prometheus.GaugeValue, float64(spod.FreeSpace), labelValues...,
		))
	}
}
