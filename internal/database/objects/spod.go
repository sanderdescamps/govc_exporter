package objects

import "time"

type StoragePod struct {
	Timestamp  time.Time               `json:"timestamp"`
	Self       ManagedObjectReference  `json:"self"`
	Parent     *ManagedObjectReference `json:"parent"`
	Name       string                  `json:"name"`
	Datacenter string                  `json:"datacenter"`

	Capacity  float64 `json:"capacity"`
	FreeSpace float64 `json:"free_space"`
}
