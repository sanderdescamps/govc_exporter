package objects

import "time"

type ResourcePool struct {
	Timestamp  time.Time               `json:"timestamp"`
	Self       ManagedObjectReference  `json:"self"`
	Parent     *ManagedObjectReference `json:"parent"`
	Name       string                  `json:"name"`
	Datacenter string                  `json:"datacenter"`

	OverallCPUUsage              float64 `json:"overall_cpu_usage"`
	OverallCPUDemand             float64 `json:"overall_cpu_demand"`
	GuestMemoryUsage             float64 `json:"guest_memory_usage"`
	HostMemoryUsage              float64 `json:"host_memory_usage"`
	DistributedCPUEntitlement    float64 `json:"distributed_cpu_entitlement"`
	DistributedMemoryEntitlement float64 `json:"distributed_memory_entitlement"`
	StaticCPUEntitlement         float64 `json:"static_cpu_entitlement"`
	PrivateMemory                float64 `json:"private_memory"`
	SharedMemory                 float64 `json:"shared_memory"`
	SwappedMemory                float64 `json:"swapped_memory"`
	BalloonedMemory              float64 `json:"ballooned_memory"`
	OverheadMemory               float64 `json:"overhead_memory"`
	ConsumedOverheadMemory       float64 `json:"consumed_overhead_memory"`
	CompressedMemory             float64 `json:"compressed_memory"`
	MemoryAllocationLimit        float64 `json:"memory_allocation_limit"`
	CPUAllocationLimit           float64 `json:"cpu_allocation_limit"`
	OverallStatus                string  `json:"overall_status"`
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
