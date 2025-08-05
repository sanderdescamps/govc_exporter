package objects

import (
	"strings"
	"time"
)

type Host struct {
	Timestamp     time.Time               `json:"timestamp"`
	Self          ManagedObjectReference  `json:"self"`
	Parent        *ManagedObjectReference `json:"parent"`
	Name          string                  `json:"name"`
	GuestHostname string                  `json:"guest_hostname"`
	Cluster       string                  `json:"cluster"`
	Datacenter    string                  `json:"datacenter"`
	Pool          string                  `json:"pool"`

	OSVersion   string `json:"os_version"`
	AssetTag    string `json:"asset_tag"`
	ServiceTag  string `json:"service_tag"`
	Vendor      string `json:"vendor"`
	Model       string `json:"model"`
	BiosVersion string `json:"bios_version"`

	PowerState                 string                    `json:"power_state"`
	ConnectionState            string                    `json:"connection_state"`
	Maintenance                bool                      `json:"maintenance"`
	UptimeSeconds              float64                   `json:"uptime_seconds"`
	RebootRequired             bool                      `json:"reboot_required"`
	CPUCoresTotal              float64                   `json:"cpu_cores_total"`
	CPUThreadsTotal            float64                   `json:"cpu_threads_total"`
	AvailCPUMhz                float64                   `json:"avail_cpu_mhz"`
	UsedCPUMhz                 float64                   `json:"used_cpu_mhz"`
	AvailMemBytes              float64                   `json:"avail_mem_bytes"`
	UsedMemBytes               float64                   `json:"used_mem_bytes"`
	OverallStatus              string                    `json:"overall_status"`
	Info                       float64                   `json:"info"`
	SystemHealthNumericSensors []HostNumericSensorHealth `json:"system_health_numeric_sensor"`
	HardwareStatus             []HardwareStatus          `json:"hardware_status"`

	GenericHBA []HostBusAdapter      `json:"generic_hba"`
	IscsiHBA   []IscsiHostBusAdapter `json:"iscsi_hba"`

	SCSILuns            []ScsiLun   `json:"scsi_lun"`
	IscsiMultiPathPaths []IscsiPath `json:"iscsi_multi_path_paths"`

	SCSILunMounted    float64 `json:"scsi_lun_mounted"`
	SCSILunAccessible float64 `json:"scsi_lun_accessible"`

	NumberOfVMs        float64 `json:"number_of_vms"`
	NumberOfDatastores float64 `json:"number_of_datastores"`
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
	Name     string `json:"name"`
	Driver   string `json:"driver"`
	Model    string `json:"model"`
	State    string `json:"state"`
	Protocol string `json:"protocol"`
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
	IQN             string                 `json:"iqn"`
	DiscoveryTarget []IscsiDiscoveryTarget `json:"send_target"`
	StaticTarget    []IscsiStaticTarget    `json:"static_target"`
}

type IscsiDiscoveryTarget struct {
	Address string `json:"address"`
	Port    int32  `json:"port"`
}

type IscsiStaticTarget struct {
	Address         string `json:"address"`
	Port            int32  `json:"port"`
	IQN             string `json:"iqn"`
	DiscoveryMethod string `json:"discovery_method"`
}

type HostNumericSensorHealth struct {
	Name  string  `json:"name"`
	Type  string  `json:"type"`
	ID    string  `json:"id"`
	Unit  string  `json:"unit"`
	Value float64 `json:"value"`
	State string  `json:"state"`
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
	Name    string `json:"name"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
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
	CanonicalName string `json:"canonical_name"`
	Vendor        string `json:"vendor"`
	Model         string `json:"model"`
	Datastore     string `json:"datastore"`
	Accessible    bool   `json:"accessible"`
	Mounted       bool   `json:"mounted"`
}

type IscsiPath struct {
	Device        string `json:"device"`
	NAA           string `json:"naa"`
	DatastoreName string `json:"datastore_name"`
	TargetAddress string `json:"target_address"`
	TargetIQN     string `json:"target_iqn"`
	State         string `json:"state"`
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
		return 3.0
	}
	return 0
}
