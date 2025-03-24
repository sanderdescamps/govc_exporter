package collector

import (
	"github.com/prometheus/client_golang/prometheus"
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

func NewClusterCollector(scraper *scraper.VCenterScraper, cConf CollectorConfig) *clusterCollector {
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
	if c.scraper.Cluster == nil {
		return
	}
	clusterData := c.scraper.Cluster.GetAllSnapshots()
	for _, snap := range clusterData {
		timestamp, p := snap.Timestamp, snap.Item
		summary := p.Summary.GetComputeResourceSummary()
		if summary == nil {
			continue
		}
		parentChain := c.scraper.GetParentChain(p.Self)
		mb := int64(1024 * 1024)

		extraLabelValues := func() []string {
			result := []string{}

			for _, tagCat := range c.extraLabels {
				tag := c.scraper.Tags.GetTag(p.Self, tagCat)
				if tag != nil {
					result = append(result, tag.Name)
				} else {
					result = append(result, "")
				}
			}
			return result
		}()

		labelValues := []string{me2id(p.ManagedEntity), p.Name, parentChain.DC}
		labelValues = append(labelValues, extraLabelValues...)

		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.totalCPU, prometheus.GaugeValue, float64(summary.TotalCpu), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.effectiveCPU, prometheus.GaugeValue, float64(summary.EffectiveCpu), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.totalMemory, prometheus.GaugeValue, float64(summary.TotalMemory), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.effectiveMemory, prometheus.GaugeValue, float64(summary.EffectiveMemory*mb), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.numCPUCores, prometheus.GaugeValue, float64(summary.NumCpuCores), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.numCPUThreads, prometheus.GaugeValue, float64(summary.NumCpuThreads), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.numEffectiveHosts, prometheus.GaugeValue, float64(summary.NumEffectiveHosts), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.numHosts, prometheus.GaugeValue, float64(summary.NumHosts), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, ConvertManagedEntityStatusToValue(summary.OverallStatus), labelValues...,
		))
	}

}
