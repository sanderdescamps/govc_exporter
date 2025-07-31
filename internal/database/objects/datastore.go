package objects

import (
	"strings"
	"time"
)

type Datastore struct {
	Timestamp        time.Time                `json:"timestamp"`
	Self             ManagedObjectReference   `json:"self"`
	Parent           *ManagedObjectReference  `json:"parent"`
	Name             string                   `json:"name"`
	Cluster          string                   `json:"cluster"`
	Kind             string                   `json:"kind"`
	Capacity         float64                  `json:"capacity"`
	FreeSpace        float64                  `json:"free_space"`
	Accessible       bool                     `json:"accessible"`
	Maintenance      string                   `json:"maintenance"`
	OverallStatus    float64                  `json:"overall_status"`
	HostAccessible   float64                  `json:"host_accessible"`
	HostMounted      float64                  `json:"host_mounted"`
	HostVmknicActive float64                  `json:"host_vmknic_active"`
	HostMountInfo    []DatastoreHostMountInfo `json:"host_mount_info"`
	VmfsInfo         *DatastoreVmfsInfo       `json:"vmfs_info"`
}

// Return the maintenance status as a float64 number.
// 0=> running, not in maintenance
// 1=> entering maintenance
// 2=> in maintenance
func (d Datastore) MaintenanceStatus() float64 {
	if strings.EqualFold(d.Maintenance, "enteringMaintenance") {
		return 1.0
	} else if strings.EqualFold(d.Maintenance, "inMaintenance") {
		return 2.0
	}
	return 0
}

type DatastoreVmfsInfo struct {
	Name  string `json:"name"`
	UUID  string `json:"uuid"`
	SSD   bool   `json:"ssd"`
	Local bool   `json:"local"`
	NAA   string `json:"naa"`
}

type DatastoreHostMountInfo struct {
	Host            string `json:"host"`
	HostID          string `json:"host_id"`
	SSD             bool   `json:"ssd"`
	Accessable      bool   `json:"accessable"`
	Mounted         bool   `json:"mounted"`
	VmknicActiveNic bool   `json:"vmknic_active_nic"`
}
