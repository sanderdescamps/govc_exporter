package collector

import (
	"context"
	"slices"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

const (
	virtualMachineCollectorSubsystem = "vm"
)

type virtualMachineCollector struct {
	scraper                *scraper.VCenterScraper
	legacyMetrics          bool
	advancedStorageMetrics bool
	advancedNetworkMetrics bool
	useIsecSpecifics       bool
	extraLabels            []string

	numCPU                      *prometheus.Desc
	numCoresPerSocket           *prometheus.Desc
	maxCPUUsage                 *prometheus.Desc
	overallCPUUsage             *prometheus.Desc
	overallCPUDemand            *prometheus.Desc
	cpuAllocationShares         *prometheus.Desc
	cpuAllocationReservation    *prometheus.Desc
	memoryBytes                 *prometheus.Desc
	guestMemoryUsage            *prometheus.Desc
	hostMemoryUsage             *prometheus.Desc
	memoryAllocationShares      *prometheus.Desc
	memoryAllocationReservation *prometheus.Desc

	uptimeSeconds        *prometheus.Desc
	numSnapshot          *prometheus.Desc
	powerState           *prometheus.Desc
	overallStatus        *prometheus.Desc
	guestHeartbeatStatus *prometheus.Desc
	toolsStatus          *prometheus.Desc
	vmInfo               *prometheus.Desc
	hostInfo             *prometheus.Desc

	// legacy metrics
	distributedCPUEntitlement    *prometheus.Desc
	distributedMemoryEntitlement *prometheus.Desc
	staticCPUEntitlement         *prometheus.Desc
	staticMemoryEntitlement      *prometheus.Desc
	privateMemory                *prometheus.Desc
	sharedMemory                 *prometheus.Desc
	swappedMemory                *prometheus.Desc
	balloonedMemory              *prometheus.Desc
	consumedOverheadMemory       *prometheus.Desc
	ftLogBandwidth               *prometheus.Desc
	ftSecondaryLatency           *prometheus.Desc
	compressedMemory             *prometheus.Desc
	ssdSwappedMemory             *prometheus.Desc

	// Advanced network metrics
	networkConnected *prometheus.Desc
	// ethernetDriverConnected *prometheus.Desc

	// Advanced storage metrics
	diskCapacityBytes *prometheus.Desc
}

func NewVirtualMachineCollector(scraper *scraper.VCenterScraper, cConf config.CollectorConfig) *virtualMachineCollector {
	labels := []string{"uuid", "name", "template", "vm_id", "pool"}
	extraLabels := cConf.VMTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	if cConf.UseIsecSpecifics {
		labels = append(labels, "crit", "responsable", "service")
	}
	infoLabels := append(slices.Clone(labels), "guest_id", "tools_version")
	hostLabels := append(slices.Clone(labels), "datacenter", "cluster", "esx")
	diskLabels := append(slices.Clone(labels), "disk_uuid", "thin_provisioned")
	networkLabels := append(slices.Clone(labels), "mac", "ip")

	return &virtualMachineCollector{
		scraper:                scraper,
		extraLabels:            extraLabels,
		legacyMetrics:          cConf.VMLegacyMetrics,
		advancedNetworkMetrics: cConf.VMAdvancedNetworkMetrics,
		advancedStorageMetrics: cConf.VMAdvancedStorageMetrics,

		numCPU: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "cpu_number"),
			"vm number of cpu", labels, nil),
		numCoresPerSocket: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "cores_per_socket"),
			"vm number of cores by socket", labels, nil),
		maxCPUUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "max_cpu_usage_mhz"),
			"total assigned CPU in MHz. Based on the host the vm is current running on", labels, nil),
		overallCPUUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "overall_cpu_usage_mhz"),
			"vm overall CPU usage in MHz", labels, nil),
		overallCPUDemand: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "overall_cpu_demand_mhz"),
			"vm overall CPU demand in MHz", labels, nil),
		cpuAllocationShares: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "cpu_allocation_shares_mhz"),
			"The number of shares allocated in MHz. ", labels, nil),
		cpuAllocationReservation: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "cpu_allocation_reservation_mhz"),
			"Amount of resource that is guaranteed available to the virtual machine or resource pool.", labels, nil),
		memoryBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "memory_bytes"),
			"vm memory in bytes", labels, nil),
		guestMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "guest_memory_usage_bytes"),
			"vm guest memory usage in bytes", labels, nil),
		hostMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "host_memory_usage_bytes"),
			"vm host memory usage in bytes", labels, nil),
		memoryAllocationShares: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "memory_allocation_shares_bytes"),
			"The number of shares allocated in bytes.", labels, nil),
		memoryAllocationReservation: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "memory_allocation_reservation_bytes"),
			"Amount of resource that is guaranteed available to the virtual machine or resource pool.", labels, nil),
		uptimeSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "uptime_seconds"),
			"vm uptime in seconds", labels, nil),
		numSnapshot: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "snapshot_number_total"),
			"vm number of snapshot", labels, nil),
		powerState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "power_state"),
			"vm power state", labels, nil),
		overallStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "overall_status"),
			"overall health status", labels, nil),
		guestHeartbeatStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "guest_heartbeat_status"),
			"Guest hartbeat status", labels, nil),
		toolsStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "tools_status"),
			"vmware tools status", labels, nil),
		vmInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "info"),
			"Info about vm", infoLabels, nil),
		hostInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "host_info"),
			"Info about the host", hostLabels, nil),

		// Advanced network metrics
		networkConnected: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "network_connected"),
			"vm network connected", networkLabels, nil),

		// Advanced storage metrics
		diskCapacityBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "disk_capacity_bytes"),
			"vm disk capacity bytes", diskLabels, nil),

		//Legacy metrics
		distributedCPUEntitlement: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "distributed_cpu_entitlement_mhz"),
			"vm distributed CPU entitlement in MHz", labels, nil),
		distributedMemoryEntitlement: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "distributed_memory_entitlement_bytes"),
			"vm distributed memory entitlement in bytes", labels, nil),
		staticCPUEntitlement: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "static_cpu_entitlement_mhz"),
			"vm static CPU entitlement in MHz", labels, nil),
		staticMemoryEntitlement: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "static_memory_entitlement_bytes"),
			"vm static memory entitlement in bytes", labels, nil),
		privateMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "private_memory_bytes"),
			"vm private memory in bytes", labels, nil),
		sharedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "shared_memory_bytes"),
			"vm shared memory in bytes", labels, nil),
		swappedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "swapped_memory_bytes"),
			"vm swapped memory in bytes", labels, nil),
		balloonedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "ballooned_memory_bytes"),
			"vm ballooned memory in bytes", labels, nil),
		consumedOverheadMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "consumed_overhead_memory_bytes"),
			"vm consumed overhead memory bytes", labels, nil),
		ftLogBandwidth: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "ft_log_bandwidth"),
			"vm ft log bandwidth", labels, nil),
		ftSecondaryLatency: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "ft_secondary_latency"),
			"vm ft secondary latency", labels, nil),
		compressedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "compressed_memory_bytes"),
			"vm compressed memory in bytes", labels, nil),
		ssdSwappedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "ssd_swapped_memory_bytes"),
			"vm ssd swapped memory in bytes", labels, nil),
	}
}

func (c *virtualMachineCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.numCPU
	ch <- c.numCoresPerSocket
	ch <- c.maxCPUUsage
	ch <- c.overallCPUUsage
	ch <- c.overallCPUDemand
	ch <- c.cpuAllocationShares
	ch <- c.cpuAllocationReservation
	ch <- c.memoryBytes
	ch <- c.guestMemoryUsage
	ch <- c.hostMemoryUsage
	ch <- c.memoryAllocationShares
	ch <- c.memoryAllocationReservation
	ch <- c.uptimeSeconds
	ch <- c.numSnapshot
	ch <- c.powerState
	ch <- c.overallStatus
	ch <- c.guestHeartbeatStatus
	ch <- c.toolsStatus
	ch <- c.vmInfo
	ch <- c.hostInfo

	// Legacy metrics
	ch <- c.distributedCPUEntitlement
	ch <- c.distributedMemoryEntitlement
	ch <- c.staticCPUEntitlement
	ch <- c.staticMemoryEntitlement
	ch <- c.privateMemory
	ch <- c.sharedMemory
	ch <- c.swappedMemory
	ch <- c.balloonedMemory
	ch <- c.consumedOverheadMemory
	ch <- c.ftLogBandwidth
	ch <- c.ftSecondaryLatency
	ch <- c.compressedMemory
	ch <- c.ssdSwappedMemory

	// Advanced network metrics
	ch <- c.networkConnected
	// Advanced storage metrics
	ch <- c.diskCapacityBytes
}

func (c *virtualMachineCollector) Collect(ch chan<- prometheus.Metric) {
	if !c.scraper.VM.Enabled() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), COLLECT_TIMEOUT)
	defer cancel()

	vms, err := c.scraper.DB.GetAllVM(ctx)
	if err != nil && Logger != nil {
		Logger.Error("failed to get vm's", "err", err)
	}
	for _, vm := range vms {
		extraLabelValues := []string{}
		objectTags := c.scraper.DB.GetTags(ctx, vm.Self)
		for _, tagCat := range c.extraLabels {
			extraLabelValues = append(extraLabelValues, objectTags.GetTag(tagCat))
		}

		labelValues := []string{vm.UUID, vm.Name, strconv.FormatBool(vm.Template), vm.Self.Value, vm.ResourcePool}
		labelValues = append(slices.Clone(labelValues), extraLabelValues...)

		if c.useIsecSpecifics && vm.IsecAnnotation != nil {
			annotation := vm.IsecAnnotation
			labelValues = append(
				slices.Clone(labelValues),
				annotation.Criticality,
				annotation.Responsable,
				annotation.Service,
			)
		}

		hostLabelValues := append(slices.Clone(labelValues), vm.Datacenter, vm.HostInfo.Cluster, vm.HostInfo.Host)

		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.numCPU, prometheus.GaugeValue, vm.NumCPU, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.numCoresPerSocket, prometheus.GaugeValue, vm.NumCoresPerSocket, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.maxCPUUsage, prometheus.GaugeValue, vm.MaxCPUUsage, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.overallCPUUsage, prometheus.GaugeValue, vm.OverallCPUUsage, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.overallCPUDemand, prometheus.GaugeValue, vm.OverallCPUDemand, labelValues...,
		))
		if vm.CPUAllocationShares != 0.0 {
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.cpuAllocationShares, prometheus.GaugeValue, vm.CPUAllocationShares, labelValues...,
			))
		}
		if vm.CPUAllocationReservation != 0.0 {
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.cpuAllocationReservation, prometheus.GaugeValue, vm.CPUAllocationReservation, labelValues...,
			))
		}
		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.memoryBytes, prometheus.GaugeValue, vm.MemoryBytes, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.guestMemoryUsage, prometheus.GaugeValue, vm.GuestMemoryUsage, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.hostMemoryUsage, prometheus.GaugeValue, vm.HostMemoryUsage, labelValues...,
		))
		if vm.MemoryAllocationShares != 0.0 {
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.memoryAllocationShares, prometheus.GaugeValue, vm.MemoryAllocationShares, labelValues...,
			))
		}
		if vm.MemoryAllocationReservation != 0.0 {
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.memoryAllocationReservation, prometheus.GaugeValue, vm.MemoryAllocationReservation, labelValues...,
			))
		}
		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.uptimeSeconds, prometheus.GaugeValue, vm.UptimeSeconds, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.numSnapshot, prometheus.GaugeValue, float64(len(vm.Snapshot)), labelValues...,
		))

		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.powerState, prometheus.GaugeValue, vm.PowerStateFloat64(), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, vm.OverallStatusFloat64(), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.guestHeartbeatStatus, prometheus.GaugeValue, vm.GuestHeartbeatStateFloat64(), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.toolsStatus, prometheus.GaugeValue, vm.GuestToolsStatusFloat64(), labelValues...,
		))

		infoLabelValues := append(slices.Clone(labelValues), vm.GuestID, vm.GuestToolsVersion)
		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.vmInfo, prometheus.GaugeValue, 0, infoLabelValues...,
		))

		ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
			c.hostInfo, prometheus.GaugeValue, 1, hostLabelValues...,
		))

		//Legacy metrics
		if c.legacyMetrics {
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.distributedCPUEntitlement, prometheus.GaugeValue, vm.DistributedCPUEntitlement, labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.distributedMemoryEntitlement, prometheus.GaugeValue, vm.DistributedMemoryEntitlement, labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.staticCPUEntitlement, prometheus.GaugeValue, vm.StaticCPUEntitlement, labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.staticMemoryEntitlement, prometheus.GaugeValue, vm.StaticMemoryEntitlement, labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.privateMemory, prometheus.GaugeValue, vm.PrivateMemory, labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.sharedMemory, prometheus.GaugeValue, vm.SharedMemory, labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.swappedMemory, prometheus.GaugeValue, vm.SwappedMemory, labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.balloonedMemory, prometheus.GaugeValue, vm.BalloonedMemory, labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.consumedOverheadMemory, prometheus.GaugeValue, vm.ConsumedOverheadMemory, labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.ftLogBandwidth, prometheus.GaugeValue, vm.FtLogBandwidth, labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.ftSecondaryLatency, prometheus.GaugeValue, vm.FtSecondaryLatency, labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.compressedMemory, prometheus.GaugeValue, vm.CompressedMemory, labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
				c.ssdSwappedMemory, prometheus.GaugeValue, vm.SsdSwappedMemory, labelValues...,
			))
		}

		// Advanced network metrics
		if c.advancedNetworkMetrics {
			for _, net := range vm.GuestNetwork {
				networkLabelValues := append(slices.Clone(labelValues), net.MacAddress, net.IpAddress)
				ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
					c.networkConnected, prometheus.GaugeValue, b2f(net.Connected), networkLabelValues...,
				))
			}
		}

		//Advanced Storage metrics
		if c.advancedStorageMetrics {
			for _, disk := range vm.Disk {
				diskLabelValues := append(slices.Clone(labelValues), disk.UUID, strconv.FormatBool(disk.ThinProvisioned))
				ch <- prometheus.NewMetricWithTimestamp(vm.Timestamp, prometheus.MustNewConstMetric(
					c.diskCapacityBytes, prometheus.GaugeValue, float64(disk.Capacity), diskLabelValues...,
				))
			}
		}
	}
}
