package objects

import "time"

type Datacenter struct {
	Timestamp time.Time               `json:"timestamp"`
	Self      ManagedObjectReference  `json:"self"`
	Parent    *ManagedObjectReference `json:"parent"`
	Name      string                  `json:"name"`
}
