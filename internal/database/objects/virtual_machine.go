package objects

import (
	"strings"
	"time"
)

type VirtualMachine struct {
	Timestamp         time.Time               `json:"timestamp"`
	TimeConfigChanged time.Time               `json:"time_config_changed"`
	TimeCreated       time.Time               `json:"time_created"`
	Self              ManagedObjectReference  `json:"self"`
	Parent            *ManagedObjectReference `json:"parent"`
	Name              string                  `json:"name"`
	UUID              string                  `json:"uuid"`
	Template          bool                    `json:"template"`

	// Cluster      string `json:"cluster"`
	Datacenter   string `json:"datacenter"`
	ResourcePool string `json:"resource_pool"`

	NumCPU                      float64 `json:"num_cpu"`
	NumCoresPerSocket           float64 `json:"num_cores_per_socket"`
	MaxCPUUsage                 float64 `json:"max_cpu_usage"`
	OverallCPUUsage             float64 `json:"overall_cpu_usage"`
	OverallCPUDemand            float64 `json:"overall_cpu_demand"`
	CPUAllocationShares         float64 `json:"cpu_allocation_shares"`
	CPUAllocationReservation    float64 `json:"cpu_allocation_reservation"`
	MemoryBytes                 float64 `json:"memory_bytes"`
	GuestMemoryUsage            float64 `json:"guest_memory_usage"`
	HostMemoryUsage             float64 `json:"host_memory_usage"`
	MemoryAllocationShares      float64 `json:"memory_allocation_shares"`
	MemoryAllocationReservation float64 `json:"memory_allocation_reservation"`

	UptimeSeconds        float64 `json:"uptime_seconds"`
	NumSnapshot          float64 `json:"num_snapshot"`
	PowerState           string  `json:"power_state"`
	OverallStatus        string  `json:"overall_status"`
	GuestHeartbeatStatus string  `json:"guest_heartbeat_status"`
	GuestToolsStatus     string  `json:"guest_tools_status"`
	GuestToolsVersion    string  `json:"guest_tools_version"`
	GuestID              string  `json:"guest_id"`

	// Legacy metrics
	DistributedCPUEntitlement    float64 `json:"distributed_cpu_entitlement"`
	DistributedMemoryEntitlement float64 `json:"distributed_memory_entitlement"`
	StaticCPUEntitlement         float64 `json:"static_cpu_entitlement"`
	StaticMemoryEntitlement      float64 `json:"static_memory_entitlement"`
	PrivateMemory                float64 `json:"private_memory"`
	SharedMemory                 float64 `json:"shared_memory"`
	SwappedMemory                float64 `json:"swapped_memory"`
	BalloonedMemory              float64 `json:"ballooned_memory"`
	ConsumedOverheadMemory       float64 `json:"consumed_overhead_memory"`
	FtLogBandwidth               float64 `json:"ft_log_bandwidth"`
	FtSecondaryLatency           float64 `json:"ft_secondary_latency"`
	CompressedMemory             float64 `json:"compressed_memory"`
	SsdSwappedMemory             float64 `json:"ssd_swapped_memory"`

	// Advanced network metrics
	GuestNetwork []VirtualMachineGuestNet `json:"network"`

	Disk     []VirtualMachineDisk     `json:"disk"`
	Snapshot []VirtualMachineSnapshot `json:"snapshot"`

	HostInfo VirtualMachineHostInfo `json:"host_info"`
}

type VirtualMachineGuestNet struct {
	Network    string   `json:"network"`
	MacAddress string   `json:"mac_address"`
	IpAddress  []string `json:"ip_address"`
	Connected  bool     `json:"connected"`
}

type VirtualMachineDisk struct {
	UUID            string  `json:"uuid"`
	VMDKFile        string  `json:"vmdk_file"`
	ThinProvisioned bool    `json:"thin_provisioned"`
	Capacity        float64 `json:"capacity"`
	Used            float64 `json:"used"`
}

type VirtualMachineSnapshot struct {
	Name         string
	CreationTime time.Time
}

type VirtualMachineHostInfo struct {
	Host       string `json:"host"`
	Pool       string `json:"pool"`
	Datacenter string `json:"datacenter"`
	Cluster    string `json:"cluster"`
}

// Return OverallStatus as float64
//
//	0 => (Gray) The status is unknown.
//	1 => (Red) The entity definitely has a problem.
//	2 => (Yellow) The entity might have a problem.
//	3 => (Green) The entity is OK.
func (vm *VirtualMachine) OverallStatusFloat64() float64 {
	return ColorToFloat64(vm.OverallStatus)
}

// Return GuestHeartbeatState as float64
//
//	0 => (Gray) The status is unknown.
//	1 => (Red) The entity definitely has a problem.
//	2 => (Yellow) The entity might have a problem.
//	3 => (Green) The entity is OK.
func (vm *VirtualMachine) GuestHeartbeatStateFloat64() float64 {
	return ColorToFloat64(vm.GuestHeartbeatStatus)
}

// Return PowerState as float64
//
//	0 => status is unknown.
//	1 => poweredOff
//	2 => suspended
//	3 => poweredOn
func (vm *VirtualMachine) PowerStateFloat64() float64 {
	if strings.EqualFold(vm.PowerState, "poweredOff") {
		return 1.0
	} else if strings.EqualFold(vm.PowerState, "suspended") {
		return 2.0
	} else if strings.EqualFold(vm.PowerState, "poweredOn") {
		return 3.0
	}
	return 0
}

// Return GuestToolsStatus as float64
//
//	0 => status unknown
//	1 => tools not installed
//	2 => tools old
//	3 => tools not running
//	4 => tools ok
func (vm *VirtualMachine) GuestToolsStatusFloat64() float64 {
	if strings.EqualFold(vm.GuestToolsStatus, "toolsNotInstalled") {
		return 1.0
	} else if strings.EqualFold(vm.GuestToolsStatus, "toolsOld") {
		return 2.0
	} else if strings.EqualFold(vm.GuestToolsStatus, "toolsNotRunning") {
		return 3.0
	} else if strings.EqualFold(vm.GuestToolsStatus, "toolsOk") {
		return 4.0
	}
	return 0
}
