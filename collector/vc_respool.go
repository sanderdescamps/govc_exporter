package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/collector/scraper"
)

const (
	resourcePoolCollectorSubsystem = "respool"
)

type resourcePoolCollector struct {
	scraper                      *scraper.VCenterScraper
	extraLabels                  []string
	overallCPUUsage              *prometheus.Desc
	overallCPUDemand             *prometheus.Desc
	guestMemoryUsage             *prometheus.Desc
	hostMemoryUsage              *prometheus.Desc
	distributedCPUEntitlement    *prometheus.Desc
	distributedMemoryEntitlement *prometheus.Desc
	staticCPUEntitlement         *prometheus.Desc
	privateMemory                *prometheus.Desc
	sharedMemory                 *prometheus.Desc
	swappedMemory                *prometheus.Desc
	balloonedMemory              *prometheus.Desc
	overheadMemory               *prometheus.Desc
	consumedOverheadMemory       *prometheus.Desc
	compressedMemory             *prometheus.Desc
	memoryLimit                  *prometheus.Desc
	overallStatus                *prometheus.Desc
}

func NewResourcePoolCollector(scraper *scraper.VCenterScraper, cConf CollectorConfig) *resourcePoolCollector {
	labels := []string{"id", "name", "datacenter"}

	extraLabels := cConf.ResourcePoolTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}
	return &resourcePoolCollector{
		scraper:     scraper,
		extraLabels: extraLabels,
		overallCPUUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "used_cpu_mhz"),
			"resource pool overall CPU usage MHz", labels, nil),
		overallCPUDemand: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "demanded_cpu_mhz"),
			"resource pool overall CPU demand MHz", labels, nil),
		guestMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "guest_used_mem_bytes"),
			"resource pool guest memory usage in bytes", labels, nil),
		hostMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "host_used_mem_bytes"),
			"resource pool host memory usage in bytes", labels, nil),
		distributedCPUEntitlement: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "distributed_cpu_entitlement_mhz"),
			"resource pool distributed CPU entitlement", labels, nil),
		distributedMemoryEntitlement: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "distributed_mem_entitlement_bytes"),
			"resource pool distributed memory entitlement", labels, nil),
		staticCPUEntitlement: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "static_cpu_entitlement_mhz"),
			"resource pool static cpu entitlement", labels, nil),
		privateMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "private_mem_bytes"),
			"resource pool private memory in bytes", labels, nil),
		sharedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "shared_mem_bytes"),
			"resource pool shared memory in bytes", labels, nil),
		swappedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "swapped_mem_bytes"),
			"resource pool swapped memory in bytes", labels, nil),
		balloonedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "ballooned_mem_bytes"),
			"resource pool ballooned memory in bytes", labels, nil),
		overheadMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "overhead_mem_bytes"),
			"resource pool overhead memory in bytes", labels, nil),
		consumedOverheadMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "consumed_overhead_mem_bytes"),
			"resource pool consumed overhead memory in bytes", labels, nil),
		compressedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "compressed_mem_bytes"),
			"resource pool compressed memory in bytes", labels, nil),
		memoryLimit: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "mem_limit_bytes"),
			"resource pool memory limit in bytes", labels, nil),
		overallStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "overall_status"),
			"overall health status", labels, nil),
	}
}

func (c *resourcePoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.overallCPUUsage
	ch <- c.overallCPUDemand
	ch <- c.guestMemoryUsage
	ch <- c.hostMemoryUsage
	ch <- c.distributedCPUEntitlement
	ch <- c.distributedMemoryEntitlement
	ch <- c.staticCPUEntitlement
	ch <- c.privateMemory
	ch <- c.sharedMemory
	ch <- c.swappedMemory
	ch <- c.balloonedMemory
	ch <- c.overheadMemory
	ch <- c.consumedOverheadMemory
	ch <- c.compressedMemory
	ch <- c.memoryLimit
	ch <- c.overallStatus
}

func (c *resourcePoolCollector) Collect(ch chan<- prometheus.Metric) {
	if c.scraper.ResourcePool == nil {
		return
	}
	resourcePoolData := c.scraper.ResourcePool.GetAllSnapshots()
	for _, snap := range resourcePoolData {
		timestamp, p := snap.Timestamp, snap.Item

		summary := p.Summary.GetResourcePoolSummary()
		if summary == nil || summary.QuickStats == nil {
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
			c.overallCPUUsage, prometheus.GaugeValue, float64(summary.QuickStats.OverallCpuUsage), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.overallCPUDemand, prometheus.GaugeValue, float64(summary.QuickStats.OverallCpuDemand), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.guestMemoryUsage, prometheus.GaugeValue, float64(summary.QuickStats.GuestMemoryUsage*mb), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.hostMemoryUsage, prometheus.GaugeValue, float64(summary.QuickStats.HostMemoryUsage*mb), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.distributedCPUEntitlement, prometheus.GaugeValue, float64(summary.QuickStats.DistributedCpuEntitlement), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.distributedMemoryEntitlement, prometheus.GaugeValue, float64(summary.QuickStats.DistributedMemoryEntitlement*mb), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.staticCPUEntitlement, prometheus.GaugeValue, float64(summary.QuickStats.StaticCpuEntitlement), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.privateMemory, prometheus.GaugeValue, float64(summary.QuickStats.PrivateMemory*mb), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.sharedMemory, prometheus.GaugeValue, float64(summary.QuickStats.SharedMemory*mb), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.swappedMemory, prometheus.GaugeValue, float64(summary.QuickStats.SwappedMemory*mb), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.balloonedMemory, prometheus.GaugeValue, float64(summary.QuickStats.BalloonedMemory*mb), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.overheadMemory, prometheus.GaugeValue, float64(summary.QuickStats.OverheadMemory*mb), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.consumedOverheadMemory, prometheus.GaugeValue, float64(summary.QuickStats.ConsumedOverheadMemory*mb), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.compressedMemory, prometheus.GaugeValue, float64(summary.QuickStats.CompressedMemory*mb), labelValues...,
		))
		if summary.Config.MemoryAllocation.Limit != nil {
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.memoryLimit, prometheus.GaugeValue, float64(*summary.Config.MemoryAllocation.Limit*mb), labelValues...,
			))
		} else {
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.memoryLimit, prometheus.GaugeValue, float64(*summary.Config.MemoryAllocation.Limit*mb), labelValues...,
			))
		}
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, ConvertManagedEntityStatusToValue(p.OverallStatus), labelValues...,
		))

	}

}
