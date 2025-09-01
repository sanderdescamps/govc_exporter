package objects

import (
	"strings"
	"time"
)

type Host struct {
	Timestamp  time.Time               `json:"timestamp" redis:"timestamp"`
	Self       ManagedObjectReference  `json:"self" redis:"self"`
	Parent     *ManagedObjectReference `json:"parent" redis:"parent"`
	Name       string                  `json:"name" redis:"name"`
	Cluster    string                  `json:"cluster" redis:"cluster"`
	Datacenter string                  `json:"datacenter" redis:"datacenter"`

	OSVersion   string `json:"os_version" redis:"os_version"`
	AssetTag    string `json:"asset_tag" redis:"asset_tag"`
	ServiceTag  string `json:"service_tag" redis:"service_tag"`
	Vendor      string `json:"vendor" redis:"vendor"`
	Model       string `json:"model" redis:"model"`
	BiosVersion string `json:"bios_version" redis:"bios_version"`

	PowerState                 string                    `json:"power_state" redis:"power_state"`
	ConnectionState            string                    `json:"connection_state" redis:"connection_state"`
	Maintenance                bool                      `json:"maintenance" redis:"maintenance"`
	UptimeSeconds              float64                   `json:"uptime_seconds" redis:"uptime_seconds"`
	RebootRequired             bool                      `json:"reboot_required" redis:"reboot_required"`
	CPUCoresTotal              float64                   `json:"cpu_cores_total" redis:"cpu_cores_total"`
	CPUThreadsTotal            float64                   `json:"cpu_threads_total" redis:"cpu_threads_total"`
	AvailCPUMhz                float64                   `json:"avail_cpu_mhz" redis:"avail_cpu_mhz"`
	UsedCPUMhz                 float64                   `json:"used_cpu_mhz" redis:"used_cpu_mhz"`
	AvailMemBytes              float64                   `json:"avail_mem_bytes" redis:"avail_mem_bytes"`
	UsedMemBytes               float64                   `json:"used_mem_bytes" redis:"used_mem_bytes"`
	OverallStatus              string                    `json:"overall_status" redis:"overall_status"`
	SystemHealthNumericSensors []HostNumericSensorHealth `json:"system_health_numeric_sensor" redis:"system_health_numeric_sensor"`
	HardwareStatus             []HardwareStatus          `json:"hardware_status" redis:"hardware_status"`

	Volumes           []Volume            `json:"volume" redis:"volume"`
	Luns              []SCSILun           `json:"luns" redis:"luns"`
	HBA               []HostBusAdapter    `json:"hba" redis:"hba"`
	MultipathPathInfo []MultipathPathInfo `json:"multipath_path_info" redis:"multipath_path_info"`

	// SCSILunMounted    float64 `json:"scsi_lun_mounted" redis:"scsi_lun_mounted"`
	// SCSILunAccessible float64 `json:"scsi_lun_accessible" redis:"scsi_lun_accessible"`

	NumberOfVMs        float64 `json:"number_of_vms" redis:"number_of_vms"`
	NumberOfDatastores float64 `json:"number_of_datastores" redis:"number_of_datastores"`
}

// Return OverallStatus as float64
//
//	0 => (Gray) The status is unknown.
//	1 => (Red) The entity definitely has a problem.
//	2 => (Yellow) The entity might have a problem.
//	3 => (Green) The entity is OK.
func (h *Host) OverallStatusFloat64() float64 {
	return ColorToFloat64(h.OverallStatus)
}

// Return PowerState as float64
//
//	0 => unknown
//	1 => poweredOn
//	2 => poweredOff
//	3 => standBy
func (h *Host) PowerStateFloat64() float64 {
	if strings.EqualFold(h.PowerState, "poweredOn") {
		return 1.0
	} else if strings.EqualFold(h.PowerState, "poweredOff") {
		return 2.0
	} else if strings.EqualFold(h.PowerState, "standBy") {
		return 3.0
	}
	return 0
}

// Return ConnectionState as float64
//
//	0 => Unknown
//	1 => Connected
//	2 => Not Responding
//	3 => Disconnected
func (h *Host) ConnectionStateFloat64() float64 {
	if strings.EqualFold(h.ConnectionState, "connected") {
		return 1.0
	} else if strings.EqualFold(h.ConnectionState, "notResponding") {
		return 2.0
	} else if strings.EqualFold(h.ConnectionState, "disconnected") {
		return 3.0
	}
	return 0
}

///////////////////
// SENSOR HEALTH
///////////////////

type HostNumericSensorHealth struct {
	Name  string  `json:"name" redis:"name"`
	Type  string  `json:"type" redis:"type"`
	ID    string  `json:"id" redis:"id"`
	Unit  string  `json:"unit" redis:"unit"`
	Value float64 `json:"value" redis:"value"`
	State string  `json:"state" redis:"state"`
}

type HardwareStatus struct {
	Name    string `json:"name" redis:"name"`
	Type    string `json:"type" redis:"type"`
	Status  string `json:"status" redis:"status"`
	Summary string `json:"summary" redis:"summary"`
}

// Return numeric sensor health state as a float64 number.
//
//	0=> Unknown
//	1=> Red
//	2=> Yellow
//	3=> Green
func (h *HostNumericSensorHealth) HealthStatus() float64 {
	return ColorToFloat64(h.State)
}

// Return numeric sensor health state as a float64 number.
//
//	0 => Gray
//	1 => Red
//	2 => Yellow
//	3 => Green
func (h *HardwareStatus) HealthStatus() float64 {
	return ColorToFloat64(h.Status)
}

///////////////////
// STORAGE
///////////////////

type Volume struct {
	Name       string `json:"name" redis:"name"`
	Type       string `json:"type" redis:"type"`
	Path       string `json:"path" redis:"path"`
	UUID       string `json:"uuid" redis:"uuid"`
	Capacity   int64  `json:"capacity" redis:"capacity"`
	Mounted    bool   `json:"mounted" redis:"mounted"`
	Accessible bool   `json:"accessible" redis:"accessible"`
	AccessMode string `json:"access_mode" redis:"access_mode"`
	DiskName   string `json:"disk_name" redis:"disk_name"`
	SSD        bool   `json:"ssd" redis:"ssd"`
	Local      bool   `json:"local" redis:"local"`
}

type SCSILun struct {
	CanonicalName     string `json:"canonical_name" redis:"canonical_name"`
	Vendor            string `json:"vendor" redis:"vendor"`
	Model             string `json:"model" redis:"model"`
	SSD               bool   `json:"ssd" redis:"ssd"`
	Local             bool   `json:"local" redis:"local"`
	TotalNumberPaths  int64  `json:"total_number_paths" redis:"total_number_paths"`
	ActiveNumberPaths int64  `json:"active_number_paths" redis:"active_number_paths"`
}

type HostBusAdapter struct {
	Type                 string                 `json:"type" redis:"type"`
	Device               string                 `json:"device" redis:"device"`
	Status               string                 `json:"status" redis:"status"`
	Model                string                 `json:"model" redis:"model"`
	Driver               string                 `json:"driver" redis:"driver"`
	Protocol             string                 `json:"protocol" redis:"protocol"`
	IscsiInitiatorIQN    string                 `json:"iscsi_initiator_iqn" redis:"iscsi_initiator_iqn"`
	IscsiDiscoveryTarget []IscsiDiscoveryTarget `json:"iscsi_discovery_target" redis:"iscsi_discovery_target"`
	IscsiStaticTarget    []IscsiStaticTarget    `json:"iscsi_static_target" redis:"iscsi_static_target"`
}

type IscsiDiscoveryTarget struct {
	Address string `json:"address" redis:"address"`
	Port    int32  `json:"port" redis:"port"`
}

type IscsiStaticTarget struct {
	Address         string `json:"address" redis:"address"`
	Port            int32  `json:"port" redis:"port"`
	IQN             string `json:"iqn" redis:"iqn"`
	DiscoveryMethod string `json:"discovery_method" redis:"discovery_method"`
}

type MultipathPathInfo struct {
	Type    string
	Name    string
	State   string
	Adapter string
	LUN     int

	IscsiTargetAddress string
	IscsiTargetIQN     string
}

// Return state as a float64 number.
//
//	0 => Unknown
//	1 => Standby
//	2 => Active
//	3 => Disabled
//	4 => Dead
func (p *MultipathPathInfo) StateFloat64() float64 {
	if strings.EqualFold(p.State, "standby") {
		return 1.0
	} else if strings.EqualFold(p.State, "active") {
		return 2.0
	} else if strings.EqualFold(p.State, "disabled") {
		return 3.0
	} else if strings.EqualFold(p.State, "dead") {
		return 4.0
	}
	return 0
}

// Return numeric sensor health state as a float64 number.
// 0=> Unknown
// 1=> Unbound
// 2=> Offline
// 3=> Online
func (hba HostBusAdapter) StatusFloat64() float64 {
	if strings.EqualFold(hba.Status, "unbound") {
		return 1.0
	} else if strings.EqualFold(hba.Status, "offline") {
		return 2.0
	} else if strings.EqualFold(hba.Status, "online") {
		return 3.0
	}
	return 0
}
