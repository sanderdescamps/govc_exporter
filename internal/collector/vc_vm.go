package collector

import (
	"encoding/json"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
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

func NewVirtualMachineCollector(scraper *scraper.VCenterScraper, cConf Config) *virtualMachineCollector {
	labels := []string{"uuid", "name", "template", "vm_id"}
	extraLabels := cConf.VMTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	if cConf.UseIsecSpecifics {
		labels = append(labels, "crit", "responsable", "service")
	}
	infoLabels := append(labels, "guest_id", "tools_version")
	hostLabels := append(labels, "pool", "datacenter", "cluster", "esx")
	diskLabels := append(labels, "disk_uuid", "thin_provisioned")
	networkLabels := append(labels, "network", "mac", "ip")

	return &virtualMachineCollector{
		scraper:                scraper,
		extraLabels:            extraLabels,
		legacyMetrics:          cConf.VMLegacyMetrics,
		advancedNetworkMetrics: cConf.VMAdvancedNetworkMetrics,
		advancedStorageMetrics: cConf.VMAdvancedStorageMetrics,
		useIsecSpecifics:       cConf.UseIsecSpecifics,

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
	if c.scraper.VM == nil {
		return
	}

	vmData := c.scraper.VM.GetAllSnapshots()
	for _, snap := range vmData {
		timestamp, vm := snap.Timestamp, snap.Item
		summary := vm.Summary
		qs := summary.QuickStats
		mb := int64(1024 * 1024)

		esxName := "NONE"
		hostRef := vm.Runtime.Host
		if hostRef != nil {
			host := c.scraper.Host.Get(*hostRef)
			if host != nil {
				esxName = host.Name
			}
		}

		parentChain := c.scraper.GetParentChain(vm.Self)

		extraLabelValues := func() []string {
			result := []string{}

			for _, tagCat := range c.extraLabels {
				tag := c.scraper.Tags.GetTag(vm.Self, tagCat)
				if tag != nil {
					result = append(result, tag.Name)
				} else {
					result = append(result, "")
				}
			}
			return result
		}()

		labelValues := []string{vm.Config.Uuid, vm.Name, strconv.FormatBool(vm.Config.Template), vm.Self.Value}
		labelValues = append(labelValues, extraLabelValues...)
		if c.useIsecSpecifics {
			annotation := GetIsecAnnotation(vm)
			labelValues = append(
				labelValues,
				annotation.Criticality,
				annotation.Responsable,
				annotation.Service,
			)
		}
		hostLabelValues := append(labelValues, parentChain.ResourcePool, parentChain.DC, parentChain.Cluster, esxName)

		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.numCPU, prometheus.GaugeValue, float64(vm.Config.Hardware.NumCPU), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.numCoresPerSocket, prometheus.GaugeValue, float64(vm.Config.Hardware.NumCoresPerSocket), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.maxCPUUsage, prometheus.GaugeValue, float64(vm.Runtime.MaxCpuUsage), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.overallCPUUsage, prometheus.GaugeValue, float64(qs.OverallCpuUsage), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.overallCPUDemand, prometheus.GaugeValue, float64(qs.OverallCpuDemand), labelValues...,
		))
		if vm.Config.CpuAllocation != nil && vm.Config.CpuAllocation.Shares != nil {
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.cpuAllocationShares, prometheus.GaugeValue, float64(vm.Config.CpuAllocation.Shares.Shares), labelValues...,
			))
		}
		if vm.Config.CpuAllocation != nil && vm.Config.CpuAllocation.Reservation != nil {
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.cpuAllocationReservation, prometheus.GaugeValue, float64(*vm.Config.CpuAllocation.Reservation), labelValues...,
			))
		}
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.memoryBytes, prometheus.GaugeValue, float64(int64(vm.Config.Hardware.MemoryMB)*mb), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.guestMemoryUsage, prometheus.GaugeValue, float64(int64(qs.GuestMemoryUsage)*mb), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.hostMemoryUsage, prometheus.GaugeValue, float64(int64(qs.HostMemoryUsage)*mb), labelValues...,
		))
		if vm.Config.MemoryAllocation != nil && vm.Config.MemoryAllocation.Shares != nil {
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.memoryAllocationShares, prometheus.GaugeValue, float64(int64(vm.Config.MemoryAllocation.Shares.Shares)*mb), labelValues...,
			))
		}
		if vm.Config.MemoryAllocation != nil && vm.Config.MemoryAllocation.Reservation != nil {
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.memoryAllocationReservation, prometheus.GaugeValue, float64(int64(*vm.Config.MemoryAllocation.Reservation)*mb), labelValues...,
			))
		}
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.uptimeSeconds, prometheus.GaugeValue, float64(qs.UptimeSeconds), labelValues...,
		))
		if vm.Snapshot != nil {
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.numSnapshot, prometheus.GaugeValue, float64(len(vm.Snapshot.RootSnapshotList)), labelValues...,
			))
		}
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.powerState, prometheus.GaugeValue, ConvertVirtualMachinePowerStateToValue(vm.Runtime.PowerState), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, ConvertManagedEntityStatusToValue(vm.OverallStatus), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.guestHeartbeatStatus, prometheus.GaugeValue, ConvertManagedEntityStatusToValue(vm.Summary.QuickStats.GuestHeartbeatStatus), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.toolsStatus, prometheus.GaugeValue, ConvertVirtualMachineToolsStatusToValue(vm.Guest.ToolsStatus), labelValues...,
		))

		infoLabelValues := append(labelValues, vm.Config.GuestId, vm.Guest.ToolsVersion)
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.vmInfo, prometheus.GaugeValue, 0, infoLabelValues...,
		))

		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.hostInfo, prometheus.GaugeValue, 1, hostLabelValues...,
		))

		//Legacy metrics
		if c.legacyMetrics {
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.distributedCPUEntitlement, prometheus.GaugeValue, float64(qs.DistributedCpuEntitlement), labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.distributedMemoryEntitlement, prometheus.GaugeValue, float64(int64(qs.DistributedMemoryEntitlement)*mb), labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.staticCPUEntitlement, prometheus.GaugeValue, float64(qs.StaticCpuEntitlement), labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.staticMemoryEntitlement, prometheus.GaugeValue, float64(int64(qs.StaticMemoryEntitlement)*mb), labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.privateMemory, prometheus.GaugeValue, float64(int64(qs.PrivateMemory)*mb), labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.sharedMemory, prometheus.GaugeValue, float64(int64(qs.SharedMemory)*mb), labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.swappedMemory, prometheus.GaugeValue, float64(int64(qs.SwappedMemory)*mb), labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.balloonedMemory, prometheus.GaugeValue, float64(int64(qs.BalloonedMemory)*mb), labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.consumedOverheadMemory, prometheus.GaugeValue, float64(int64(qs.ConsumedOverheadMemory)*mb), labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.ftLogBandwidth, prometheus.GaugeValue, float64(qs.FtLogBandwidth), labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.ftSecondaryLatency, prometheus.GaugeValue, float64(qs.FtSecondaryLatency), labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.compressedMemory, prometheus.GaugeValue, float64(int64(qs.CompressedMemory)*mb), labelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
				c.ssdSwappedMemory, prometheus.GaugeValue, float64(int64(qs.SsdSwappedMemory)*mb), labelValues...,
			))
		}

		// Advanced network metrics
		if vm.Guest != nil && c.advancedNetworkMetrics {
			for _, net := range vm.Guest.Net {
				for _, ip := range net.IpAddress {
					networkLabelValues := append(labelValues, net.Network, net.MacAddress, ip)
					ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
						c.networkConnected, prometheus.GaugeValue, b2f(net.Connected), networkLabelValues...,
					))
				}
			}
		}

		//Advanced Storage metrics
		if c.advancedStorageMetrics {
			for _, disk := range GetDisks(vm) {
				diskLabelValues := append(labelValues, disk.UUID, strconv.FormatBool(disk.ThinProvisioned))
				ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
					c.diskCapacityBytes, prometheus.GaugeValue, float64(disk.Capacity), diskLabelValues...,
				))
			}
		}
	}
}

type Disk struct {
	UUID            string
	VmdkFile        string
	Capacity        int64
	ThinProvisioned bool
}

func GetDisks(vm mo.VirtualMachine) []Disk {
	disks := object.VirtualDeviceList(vm.Config.Hardware.Device).SelectByType((*types.VirtualDisk)(nil))
	res := make([]Disk, 0, len(disks))
	for _, d := range disks {
		disk := d.(*types.VirtualDisk)
		info := disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo)
		res = append(res, Disk{
			UUID:            info.Uuid,
			VmdkFile:        info.FileName,
			Capacity:        disk.CapacityInBytes,
			ThinProvisioned: *info.ThinProvisioned,
		})
	}
	return res
}

type IsecAnnotation struct {
	Criticality string `json:"crit"`
	Responsable string `json:"resp"`
	Service     string `json:"svc"`
}

func GetIsecAnnotation(vm mo.VirtualMachine) IsecAnnotation {
	tmp := IsecAnnotation{
		Service:     "not defined",
		Responsable: "not defined",
		Criticality: "not defined",
	}
	_ = json.Unmarshal([]byte(vm.Config.Annotation), &tmp)
	return tmp
}

func ConvertVirtualMachinePowerStateToValue(s types.VirtualMachinePowerState) float64 {
	if s == types.VirtualMachinePowerStatePoweredOff {
		return 1.0
	} else if s == types.VirtualMachinePowerStateSuspended {
		return 2.0
	} else if s == types.VirtualMachinePowerStatePoweredOn {
		return 3.0
	}
	return 0
}

func ConvertManagedEntityStatusToValue(s types.ManagedEntityStatus) float64 {
	if s == types.ManagedEntityStatusRed {
		return 1.0
	} else if s == types.ManagedEntityStatusYellow {
		return 2.0
	} else if s == types.ManagedEntityStatusGreen {
		return 3.0
	}
	return 0
}

func ConvertVirtualMachineGuestStateToValue(s types.VirtualMachineGuestState) float64 {
	if s == types.VirtualMachineGuestStateNotRunning {
		return 1.0
	} else if s == types.VirtualMachineGuestStateResetting {
		return 2.0
	} else if s == types.VirtualMachineGuestStateShuttingDown {
		return 3.0
	} else if s == types.VirtualMachineGuestStateStandby {
		return 4.0
	} else if s == types.VirtualMachineGuestStateRunning {
		return 5.0
	}
	return 0
}

func ConvertVirtualMachineToolsStatusToValue(s types.VirtualMachineToolsStatus) float64 {
	if s == types.VirtualMachineToolsStatusToolsNotInstalled {
		return 1.0
	} else if s == types.VirtualMachineToolsStatusToolsOld {
		return 2.0
	} else if s == types.VirtualMachineToolsStatusToolsNotRunning {
		return 3.0
	} else if s == types.VirtualMachineToolsStatusToolsOk {
		return 4.0
	}
	return 0
}
