// Copyright 2020 Intrinsec
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !noesx
// +build !noesx

package collector

import (
	"github.com/intrinsec/govc_exporter/collector/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	resourcePoolCollectorSubsystem = "respool"
)

type resourcePoolCollector struct {
	scraper                      *scraper.VCenterScraper
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

func NewResourcePoolCollector(scraper *scraper.VCenterScraper) *resourcePoolCollector {
	labels := []string{"id", "name", "datacenter"}
	return &resourcePoolCollector{
		scraper: scraper,
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
	resourcePools := c.scraper.ResourcePool.GetAll()
	for _, p := range resourcePools {
		summary := p.Summary.GetResourcePoolSummary()
		if summary == nil || summary.QuickStats == nil {
			continue
		}
		parentChain := c.scraper.GetParentChain(p.Self)
		mb := int64(1024 * 1024)
		labelValues := []string{me2id(p.ManagedEntity), p.Name, parentChain.DC}
		ch <- prometheus.MustNewConstMetric(
			c.overallCPUUsage, prometheus.GaugeValue, float64(summary.QuickStats.OverallCpuUsage), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.overallCPUDemand, prometheus.GaugeValue, float64(summary.QuickStats.OverallCpuDemand), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.guestMemoryUsage, prometheus.GaugeValue, float64(summary.QuickStats.GuestMemoryUsage*mb), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.hostMemoryUsage, prometheus.GaugeValue, float64(summary.QuickStats.HostMemoryUsage*mb), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.distributedCPUEntitlement, prometheus.GaugeValue, float64(summary.QuickStats.DistributedCpuEntitlement), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.distributedMemoryEntitlement, prometheus.GaugeValue, float64(summary.QuickStats.DistributedMemoryEntitlement*mb), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.staticCPUEntitlement, prometheus.GaugeValue, float64(summary.QuickStats.StaticCpuEntitlement), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.privateMemory, prometheus.GaugeValue, float64(summary.QuickStats.PrivateMemory*mb), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.sharedMemory, prometheus.GaugeValue, float64(summary.QuickStats.SharedMemory*mb), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.swappedMemory, prometheus.GaugeValue, float64(summary.QuickStats.SwappedMemory*mb), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.balloonedMemory, prometheus.GaugeValue, float64(summary.QuickStats.BalloonedMemory*mb), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.overheadMemory, prometheus.GaugeValue, float64(summary.QuickStats.OverheadMemory*mb), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.consumedOverheadMemory, prometheus.GaugeValue, float64(summary.QuickStats.ConsumedOverheadMemory*mb), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.compressedMemory, prometheus.GaugeValue, float64(summary.QuickStats.CompressedMemory*mb), labelValues...,
		)
		if summary.Config.MemoryAllocation.Limit != nil {
			ch <- prometheus.MustNewConstMetric(
				c.memoryLimit, prometheus.GaugeValue, float64(*summary.Config.MemoryAllocation.Limit*mb), labelValues...,
			)
		} else {
			ch <- prometheus.MustNewConstMetric(
				c.memoryLimit, prometheus.GaugeValue, float64(*summary.Config.MemoryAllocation.Limit*mb), labelValues...,
			)
		}
		ch <- prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, ConvertManagedEntityStatusToValue(p.OverallStatus), labelValues...,
		)

	}

}
