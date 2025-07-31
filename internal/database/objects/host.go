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

	OSVersion  string `json:"os_version"`
	AssetTag   string `json:"asset_tag"`
	ServiceTag string `json:"service_tag"`
	Vendor     string `json:"vendor"`
	Model      string `json:"model"`

	PowerState                     string                    `json:"power_state"`
	ConnectionState                string                    `json:"connection_state"`
	Maintenance                    bool                      `json:"maintenance"`
	UptimeSeconds                  float64                   `json:"uptime_seconds"`
	RebootRequired                 bool                      `json:"reboot_required"`
	CPUCoresTotal                  float64                   `json:"cpu_cores_total"`
	CPUThreadsTotal                float64                   `json:"cpu_threads_total"`
	AvailCPUMhz                    float64                   `json:"avail_cpu_mhz"`
	UsedCPUMhz                     float64                   `json:"used_cpu_mhz"`
	AvailMemBytes                  float64                   `json:"avail_mem_bytes"`
	UsedMemBytes                   float64                   `json:"used_mem_bytes"`
	OverallStatus                  float64                   `json:"overall_status"`
	Info                           float64                   `json:"info"`
	SystemHealthNumerivSensor      []HostNumericSensorHealth `json:"system_health_numeric_sensor"`
	SystemHealthNumericSensorState float64                   `json:"system_health_numeric_sensor_state"`
	SystemHealthStatusSensor       float64                   `json:"system_health_status_sensor"`

	GenericHBA []HostBusAdapter      `json:"generic_hba"`
	IscsiHBA   []IscsiHostBusAdapter `json:"iscsi_hba"`

	HBAStatus                float64 `json:"hba_status"`
	HBAIscsiSendTargetInfo   float64 `json:"hba_iscsi_send_target_info"`
	HBAIscsiStaticTargetInfo float64 `json:"hba_iscsi_static_target_info"`
	MultipathPathState       float64 `json:"multipath_path_state"`
	SCSILunMounted           float64 `json:"scsi_lun_mounted"`
	SCSILunAccessible        float64 `json:"scsi_lun_accessible"`

	NumberOfVMs        float64 `json:"number_of_vms"`
	NumberOfDatastores float64 `json:"number_of_datastores"`
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
func (hba HostBusAdapter) Status() float64 {
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
	if h.State == "" {
		return 0
	} else if strings.EqualFold(h.State, "Red") {
		return 1.0
	} else if strings.EqualFold(h.State, "Yellow") {
		return 2.0
	} else if strings.EqualFold(h.State, "Green") {
		return 3.0
	}
	return 0
}
