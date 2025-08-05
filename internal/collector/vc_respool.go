package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
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
	memoryAllocationLimit        *prometheus.Desc
	cpuAllocationLimit           *prometheus.Desc
	overallStatus                *prometheus.Desc
}

func NewResourcePoolCollector(scraper *scraper.VCenterScraper, cConf Config) *resourcePoolCollector {
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
		memoryAllocationLimit: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "mememory_allocation_limit_bytes"),
			"resource pool memory limit in bytes", labels, nil),
		cpuAllocationLimit: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, resourcePoolCollectorSubsystem, "cpu_allocation_limit"),
			"resource pool cpu limit", labels, nil),
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
	ch <- c.memoryAllocationLimit
	ch <- c.cpuAllocationLimit
	ch <- c.overallStatus
}

func (c *resourcePoolCollector) Collect(ch chan<- prometheus.Metric) {
	if !c.scraper.Cluster.Enabled() {
		return
	}
	ctx := context.Background()
	for rpool := range c.scraper.DB.GetAllResourcePoolIter(ctx) {
		extraLabelValues := []string{}
		objectTags := c.scraper.DB.GetTags(ctx, rpool.Self)
		for _, tagCat := range c.extraLabels {
			extraLabelValues = append(extraLabelValues, objectTags.GetTag(tagCat))
		}

		labelValues := []string{rpool.Self.ID(), rpool.Name, rpool.Datacenter}
		labelValues = append(labelValues, extraLabelValues...)

		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.overallCPUUsage, prometheus.GaugeValue, rpool.OverallCPUUsage, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.overallCPUDemand, prometheus.GaugeValue, rpool.OverallCPUDemand, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.guestMemoryUsage, prometheus.GaugeValue, rpool.GuestMemoryUsage, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.hostMemoryUsage, prometheus.GaugeValue, rpool.HostMemoryUsage, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.distributedCPUEntitlement, prometheus.GaugeValue, rpool.DistributedCPUEntitlement, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.distributedMemoryEntitlement, prometheus.GaugeValue, rpool.DistributedMemoryEntitlement, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.staticCPUEntitlement, prometheus.GaugeValue, rpool.StaticCPUEntitlement, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.privateMemory, prometheus.GaugeValue, rpool.PrivateMemory, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.sharedMemory, prometheus.GaugeValue, rpool.SharedMemory, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.swappedMemory, prometheus.GaugeValue, rpool.SwappedMemory, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.balloonedMemory, prometheus.GaugeValue, rpool.BalloonedMemory, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.overheadMemory, prometheus.GaugeValue, rpool.OverheadMemory, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.consumedOverheadMemory, prometheus.GaugeValue, rpool.ConsumedOverheadMemory, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.compressedMemory, prometheus.GaugeValue, rpool.CompressedMemory, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.memoryAllocationLimit, prometheus.GaugeValue, rpool.MemoryAllocationLimit, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.cpuAllocationLimit, prometheus.GaugeValue, rpool.CPUAllocationLimit, labelValues...,
		))

		ch <- prometheus.NewMetricWithTimestamp(rpool.Timestamp, prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, rpool.OverallStatusFloat64(), labelValues...,
		))

	}

}
