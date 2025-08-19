package objects

import (
	"strings"
	"time"
)

type Host struct {
	Timestamp     time.Time               `json:"timestamp" redis:"timestamp"`
	Self          ManagedObjectReference  `json:"self" redis:"self"`
	Parent        *ManagedObjectReference `json:"parent" redis:"parent"`
	Name          string                  `json:"name" redis:"name"`
	GuestHostname string                  `json:"guest_hostname" redis:"guest_hostname"`
	Cluster       string                  `json:"cluster" redis:"cluster"`
	Datacenter    string                  `json:"datacenter" redis:"datacenter"`
	Pool          string                  `json:"pool" redis:"pool"`

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
	Info                       float64                   `json:"info" redis:"info"`
	SystemHealthNumericSensors []HostNumericSensorHealth `json:"system_health_numeric_sensor" redis:"system_health_numeric_sensor"`
	HardwareStatus             []HardwareStatus          `json:"hardware_status" redis:"hardware_status"`

	GenericHBA []HostBusAdapter      `json:"generic_hba" redis:"generic_hba"`
	IscsiHBA   []IscsiHostBusAdapter `json:"iscsi_hba" redis:"iscsi_hba"`

	SCSILuns            []ScsiLun   `json:"scsi_lun" redis:"scsi_lun"`
	IscsiMultiPathPaths []IscsiPath `json:"iscsi_multi_path_paths" redis:"iscsi_multi_path_paths"`

	SCSILunMounted    float64 `json:"scsi_lun_mounted" redis:"scsi_lun_mounted"`
	SCSILunAccessible float64 `json:"scsi_lun_accessible" redis:"scsi_lun_accessible"`

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

// Return PowerState as float64
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

type HostBusAdapter struct {
	Name     string `json:"name" redis:"name"`
	Driver   string `json:"driver" redis:"driver"`
	Model    string `json:"model" redis:"model"`
	State    string `json:"state" redis:"state"`
	Protocol string `json:"protocol" redis:"protocol"`
}

// Return numeric sensor health state as a float64 number.
// 0=> Unknown
// 1=> Unbound
// 2=> Offline
// 3=> Online
func (hba HostBusAdapter) StatusFloat64() float64 {
	if strings.EqualFold(hba.State, "unbound") {
		return 1.0
	} else if strings.EqualFold(hba.State, "offline") {
		return 2.0
	} else if strings.EqualFold(hba.State, "online") {
		return 3.0
	}
	return 0
}

type IscsiHostBusAdapter struct {
	HostBusAdapter
	IQN             string                 `json:"iqn" redis:"iqn"`
	DiscoveryTarget []IscsiDiscoveryTarget `json:"send_target" redis:"send_target"`
	StaticTarget    []IscsiStaticTarget    `json:"static_target" redis:"static_target"`
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

type HostNumericSensorHealth struct {
	Name  string  `json:"name" redis:"name"`
	Type  string  `json:"type" redis:"type"`
	ID    string  `json:"id" redis:"id"`
	Unit  string  `json:"unit" redis:"unit"`
	Value float64 `json:"value" redis:"value"`
	State string  `json:"state" redis:"state"`
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

type HardwareStatus struct {
	Name    string `json:"name" redis:"name"`
	Type    string `json:"type" redis:"type"`
	Status  string `json:"status" redis:"status"`
	Summary string `json:"summary" redis:"summary"`
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

type ScsiLun struct {
	CanonicalName string `json:"canonical_name" redis:"canonical_name"`
	Vendor        string `json:"vendor" redis:"vendor"`
	Model         string `json:"model" redis:"model"`
	Datastore     string `json:"datastore" redis:"datastore"`
	Accessible    bool   `json:"accessible" redis:"accessible"`
	Mounted       bool   `json:"mounted" redis:"mounted"`
}

type IscsiPath struct {
	Device        string `json:"device" redis:"device"`
	NAA           string `json:"naa" redis:"naa"`
	DatastoreName string `json:"datastore_name" redis:"datastore_name"`
	TargetAddress string `json:"target_address" redis:"target_address"`
	TargetIQN     string `json:"target_iqn" redis:"target_iqn"`
	State         string `json:"state" redis:"state"`
}

// Return state as a float64 number.
//
//	0 => Unknown
//	1 => Standby
//	2 => Active
//	3 => Disabled
//	4 => Dead

func (p *IscsiPath) StateFloat64() float64 {
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
