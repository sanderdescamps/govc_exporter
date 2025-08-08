package objects

import (
	"time"
)

type Cluster struct {
	Timestamp  time.Time               `json:"timestamp" redis:"timestamp"`
	Self       ManagedObjectReference  `json:"self" redis:"self"`
	Parent     *ManagedObjectReference `json:"parent" redis:"parent"`
	Name       string                  `json:"name" redis:"name"`
	Datacenter string                  `json:"datacenter" redis:"datacenter"`

	TotalCPU          float64 `json:"total_cpu" redis:"total_cpu"`
	EffectiveCPU      float64 `json:"effective_cpu" redis:"effective_cpu"`
	TotalMemory       float64 `json:"total_memory" redis:"total_memory"`
	EffectiveMemory   float64 `json:"effective_memory" redis:"effective_memory"`
	NumCPUCores       float64 `json:"num_cpu_cores" redis:"num_cpu_cores"`
	NumCPUThreads     float64 `json:"num_cpu_threads" redis:"num_cpu_threads"`
	NumEffectiveHosts float64 `json:"num_effective_hosts" redis:"num_effective_hosts"`
	NumHosts          float64 `json:"num_hosts" redis:"num_hosts"`
	OverallStatus     string  `json:"overall_status" redis:"overall_status"`
}

// Return OverallStatus as float64
//
//	0 => (Gray) The status is unknown.
//	1 => (Red) The entity definitely has a problem.
//	2 => (Yellow) The entity might have a problem.
//	3 => (Green) The entity is OK.
func (c *Cluster) OverallStatusFloat64() float64 {
	return ColorToFloat64(c.OverallStatus)
}
