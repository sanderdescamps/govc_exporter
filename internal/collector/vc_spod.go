package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

const (
	storagePodCollectorSubsystem = "spod"
)

type storagePodCollector struct {
	scraper       *scraper.VCenterScraper
	extraLabels   []string
	capacity      *prometheus.Desc
	freeSpace     *prometheus.Desc
	overallStatus *prometheus.Desc
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
		overallStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, storagePodCollectorSubsystem, "overall_status"),
			"overall health status", labels, nil),
	}
}

func (c *storagePodCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.capacity
	ch <- c.freeSpace
	ch <- c.overallStatus
}

func (c *storagePodCollector) Collect(ch chan<- prometheus.Metric) {
	if c.scraper.SPOD == nil {
		return
	}

	storagePodData := c.scraper.SPOD.GetAllSnapshots()
	for _, snap := range storagePodData {
		timestamp, s := snap.Timestamp, snap.Item
		summary := s.Summary
		parentChain := c.scraper.GetParentChain(s.Self)

		extraLabelValues := func() []string {
			result := []string{}

			for _, tagCat := range c.extraLabels {
				tag := c.scraper.Tags.GetTag(s.Self, tagCat)
				if tag != nil {
					result = append(result, tag.Name)
				} else {
					result = append(result, "")
				}
			}
			return result
		}()

		labelValues := []string{me2id(s.ManagedEntity), s.Name, parentChain.DC}
		labelValues = append(labelValues, extraLabelValues...)
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.capacity, prometheus.GaugeValue, float64(summary.Capacity), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.freeSpace, prometheus.GaugeValue, float64(summary.FreeSpace), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, ConvertManagedEntityStatusToValue(s.OverallStatus), labelValues...,
		))
	}
}
