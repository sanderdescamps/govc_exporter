package objects

import (
	"time"
)

type Cluster struct {
	Timestamp  time.Time               `json:"timestamp"`
	Self       ManagedObjectReference  `json:"self"`
	Parent     *ManagedObjectReference `json:"parent"`
	Name       string                  `json:"name"`
	Datacenter string                  `json:"datacenter"`

	TotalCPU          float64 `json:"total_cpu"`
	EffectiveCPU      float64 `json:"effective_cpu"`
	TotalMemory       float64 `json:"total_memory"`
	EffectiveMemory   float64 `json:"effective_memory"`
	NumCPUCores       float64 `json:"num_cpu_cores"`
	NumCPUThreads     float64 `json:"num_cpu_threads"`
	NumEffectiveHosts float64 `json:"num_effective_hosts"`
	NumHosts          float64 `json:"num_hosts"`
	OverallStatus     float64 `json:"overall_status"`
}
