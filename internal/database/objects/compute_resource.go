package objects

import "time"

type ComputeResource struct {
	Timestamp time.Time               `json:"timestamp" redis:"timestamp"`
	Self      ManagedObjectReference  `json:"self" redis:"self"`
	Parent    *ManagedObjectReference `json:"parent" redis:"parent"`
	Name      string                  `json:"name" redis:"name"`
}
