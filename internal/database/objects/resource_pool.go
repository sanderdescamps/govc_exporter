package objects

import "time"

type ResourcePool struct {
	Timestamp  time.Time               `json:"timestamp" redis:"timestamp"`
	Self       ManagedObjectReference  `json:"self" redis:"self"`
	Parent     *ManagedObjectReference `json:"parent" redis:"parent"`
	Name       string                  `json:"name" redis:"name"`
	Datacenter string                  `json:"datacenter" redis:"datacenter"`

	OverallCPUUsage              float64 `json:"overall_cpu_usage" redis:"overall_cpu_usage"`
	OverallCPUDemand             float64 `json:"overall_cpu_demand" redis:"overall_cpu_demand"`
	GuestMemoryUsage             float64 `json:"guest_memory_usage" redis:"guest_memory_usage"`
	HostMemoryUsage              float64 `json:"host_memory_usage" redis:"host_memory_usage"`
	DistributedCPUEntitlement    float64 `json:"distributed_cpu_entitlement" redis:"distributed_cpu_entitlement"`
	DistributedMemoryEntitlement float64 `json:"distributed_memory_entitlement" redis:"distributed_memory_entitlement"`
	StaticCPUEntitlement         float64 `json:"static_cpu_entitlement" redis:"static_cpu_entitlement"`
	PrivateMemory                float64 `json:"private_memory" redis:"private_memory"`
	SharedMemory                 float64 `json:"shared_memory" redis:"shared_memory"`
	SwappedMemory                float64 `json:"swapped_memory" redis:"swapped_memory"`
	BalloonedMemory              float64 `json:"ballooned_memory" redis:"ballooned_memory"`
	OverheadMemory               float64 `json:"overhead_memory" redis:"overhead_memory"`
	ConsumedOverheadMemory       float64 `json:"consumed_overhead_memory" redis:"consumed_overhead_memory"`
	CompressedMemory             float64 `json:"compressed_memory" redis:"compressed_memory"`
	MemoryAllocationLimit        float64 `json:"memory_allocation_limit" redis:"memory_allocation_limit"`
	CPUAllocationLimit           float64 `json:"cpu_allocation_limit" redis:"cpu_allocation_limit"`
	OverallStatus                string  `json:"overall_status" redis:"overall_status"`
}

// Return OverallStatus as float64
//
//	0 => (Gray) The status is unknown.
//	1 => (Red) The entity definitely has a problem.
//	2 => (Yellow) The entity might have a problem.
//	3 => (Green) The entity is OK.
func (p *ResourcePool) OverallStatusFloat64() float64 {
	return ColorToFloat64(p.OverallStatus)
}
