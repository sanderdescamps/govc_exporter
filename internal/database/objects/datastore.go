package objects

import (
	"strings"
	"time"
)

type Datastore struct {
	Timestamp        time.Time                `json:"timestamp" redis:"timestamp"`
	Self             ManagedObjectReference   `json:"self" redis:"self"`
	Parent           *ManagedObjectReference  `json:"parent" redis:"parent"`
	Name             string                   `json:"name" redis:"name"`
	DatastoreCluster string                   `json:"datastore_cluster" redis:"datastore_cluster"`
	Kind             string                   `json:"kind" redis:"kind"`
	Capacity         float64                  `json:"capacity" redis:"capacity"`
	FreeSpace        float64                  `json:"free_space" redis:"free_space"`
	Accessible       bool                     `json:"accessible" redis:"accessible"`
	Maintenance      string                   `json:"maintenance" redis:"maintenance"`
	OverallStatus    string                   `json:"overall_status" redis:"overall_status"`
	HostAccessible   float64                  `json:"host_accessible" redis:"host_accessible"`
	HostMounted      float64                  `json:"host_mounted" redis:"host_mounted"`
	HostVmknicActive float64                  `json:"host_vmknic_active" redis:"host_vmknic_active"`
	HostMountInfo    []DatastoreHostMountInfo `json:"host_mount_info" redis:"host_mount_info"`
	VmfsInfo         *DatastoreVmfsInfo       `json:"vmfs_info" redis:"vmfs_info"`
}

// Return the maintenance status as a float64 number.
// 0=> running, not in maintenance
// 1=> entering maintenance
// 2=> in maintenance
func (d *Datastore) MaintenanceStatusFloat64() float64 {
	if strings.EqualFold(d.Maintenance, "enteringMaintenance") {
		return 1.0
	} else if strings.EqualFold(d.Maintenance, "inMaintenance") {
		return 2.0
	}
	return 0
}

// Return OverallStatus as float64
//
//	0 => (Gray) The status is unknown.
//	1 => (Red) The entity definitely has a problem.
//	2 => (Yellow) The entity might have a problem.
//	3 => (Green) The entity is OK.
func (d *Datastore) OverallStatusFloat64() float64 {
	return ColorToFloat64(d.OverallStatus)
}

type DatastoreVmfsInfo struct {
	Name  string `json:"name" redis:"name"`
	UUID  string `json:"uuid" redis:"uuid"`
	SSD   bool   `json:"ssd" redis:"ssd"`
	Local bool   `json:"local" redis:"local"`
	NAA   string `json:"naa" redis:"naa"`
}

type DatastoreHostMountInfo struct {
	Host            string `json:"host" redis:"host"`
	HostID          string `json:"host_id" redis:"host_id"`
	SSD             bool   `json:"ssd" redis:"ssd"`
	Accessible      bool   `json:"accessible" redis:"accessible"`
	Mounted         bool   `json:"mounted" redis:"mounted"`
	VmknicActiveNic bool   `json:"vmknic_active_nic" redis:"vmknic_active_nic"`
}
