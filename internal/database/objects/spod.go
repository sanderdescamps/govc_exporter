package objects

import "time"

type StoragePod struct {
	Timestamp  time.Time               `json:"timestamp" redis:"timestamp"`
	Self       ManagedObjectReference  `json:"self" redis:"self"`
	Parent     *ManagedObjectReference `json:"parent" redis:"parent"`
	Name       string                  `json:"name" redis:"name"`
	Datacenter string                  `json:"datacenter" redis:"datacenter"`

	Capacity  float64 `json:"capacity" redis:"capacity"`
	FreeSpace float64 `json:"free_space" redis:"free_space"`
}
