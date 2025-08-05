package objects

import "time"

type Folder struct {
	Timestamp time.Time               `json:"timestamp"`
	Self      ManagedObjectReference  `json:"self"`
	Parent    *ManagedObjectReference `json:"parent"`
	Name      string                  `json:"name"`
}
