package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

const (
	clusterCollectorSubsystem = "cluster"
)

type clusterCollector struct {
	scraper     *scraper.VCenterScraper
	extraLabels []string

	totalCPU          *prometheus.Desc
	effectiveCPU      *prometheus.Desc
	totalMemory       *prometheus.Desc
	effectiveMemory   *prometheus.Desc
	numCPUCores       *prometheus.Desc
	numCPUThreads     *prometheus.Desc
	numEffectiveHosts *prometheus.Desc
	numHosts          *prometheus.Desc
	overallStatus     *prometheus.Desc
}

func NewClusterCollector(scraper *scraper.VCenterScraper, cConf config.CollectorConfig) *clusterCollector {
	labels := []string{"id", "name", "datacenter"}

	extraLabels := cConf.ClusterTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	return &clusterCollector{
		scraper:     scraper,
		extraLabels: extraLabels,
		totalCPU: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "total_cpu_mhz"),
			"Aggregated CPU resources of all hosts, in MHz", labels, nil),
		effectiveCPU: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "effective_cpu_mhz"),
			"Effective CPU resources (in MHz) available to run virtual machines.", labels, nil),
		totalMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "total_memory"),
			"Aggregated memory resources of all hosts, in bytes", labels, nil),
		effectiveMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "effective_memory_bytes"),
			"Effective memory resources (in bytes) available to run virtual machines", labels, nil),
		numCPUCores: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "num_cpu_cores"),
			"Number of physical CPU cores. Physical CPU cores are the processors contained by a CPU package.", labels, nil),
		numCPUThreads: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "num_cpu_threads"),
			"Aggregated number of CPU threads", labels, nil),
		numEffectiveHosts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "num_effective_hosts"),
			"Total number of effective hosts", labels, nil),
		numHosts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "num_hosts"),
			"Total number of hosts", labels, nil),
		overallStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "overall_status"),
			"overall health status", labels, nil),
	}
}

func (c *clusterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalCPU
	ch <- c.effectiveCPU
	ch <- c.totalMemory
	ch <- c.effectiveMemory
	ch <- c.numCPUCores
	ch <- c.numCPUThreads
	ch <- c.numEffectiveHosts
	ch <- c.numHosts
	ch <- c.overallStatus
}

func (c *clusterCollector) Collect(ch chan<- prometheus.Metric) {
	if !c.scraper.Cluster.Enabled() {
		return
	}
	ctx := context.Background()
	for cluster := range c.scraper.DB.GetAllClusterIter(ctx) {

		extraLabelValues := []string{}
		objectTags := c.scraper.DB.GetTags(ctx, cluster.Self)
		for _, tagCat := range c.extraLabels {
			extraLabelValues = append(extraLabelValues, objectTags.GetTag(tagCat))
		}

		labelValues := []string{cluster.Self.ID(), cluster.Name, cluster.Datacenter}
		labelValues = append(labelValues, extraLabelValues...)

		ch <- prometheus.NewMetricWithTimestamp(cluster.Timestamp, prometheus.MustNewConstMetric(
			c.totalCPU, prometheus.GaugeValue, float64(cluster.TotalCPU), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(cluster.Timestamp, prometheus.MustNewConstMetric(
			c.effectiveCPU, prometheus.GaugeValue, float64(cluster.EffectiveCPU), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(cluster.Timestamp, prometheus.MustNewConstMetric(
			c.totalMemory, prometheus.GaugeValue, float64(cluster.TotalMemory), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(cluster.Timestamp, prometheus.MustNewConstMetric(
			c.effectiveMemory, prometheus.GaugeValue, float64(cluster.EffectiveMemory), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(cluster.Timestamp, prometheus.MustNewConstMetric(
			c.numCPUCores, prometheus.GaugeValue, float64(cluster.NumCPUThreads), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(cluster.Timestamp, prometheus.MustNewConstMetric(
			c.numCPUThreads, prometheus.GaugeValue, float64(cluster.NumCPUThreads), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(cluster.Timestamp, prometheus.MustNewConstMetric(
			c.numEffectiveHosts, prometheus.GaugeValue, float64(cluster.NumEffectiveHosts), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(cluster.Timestamp, prometheus.MustNewConstMetric(
			c.numHosts, prometheus.GaugeValue, cluster.NumHosts, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(cluster.Timestamp, prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, cluster.OverallStatusFloat64(), labelValues...,
		))
	}

}
