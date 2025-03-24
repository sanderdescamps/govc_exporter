package collector

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	esxCollectorSubsystem = "esx"
)

type esxCollector struct {
	// vcCollector
	enableStorageMetrics           bool
	extraLabels                    []string
	scraper                        *scraper.VCenterScraper
	powerState                     *prometheus.Desc
	connectionState                *prometheus.Desc
	maintenance                    *prometheus.Desc
	uptimeSeconds                  *prometheus.Desc
	rebootRequired                 *prometheus.Desc
	cpuCoresTotal                  *prometheus.Desc
	cpuThreadsTotal                *prometheus.Desc
	availCPUMhz                    *prometheus.Desc
	usedCPUMhz                     *prometheus.Desc
	availMemBytes                  *prometheus.Desc
	usedMemBytes                   *prometheus.Desc
	overallStatus                  *prometheus.Desc
	info                           *prometheus.Desc
	systemHealthNumericSensorValue *prometheus.Desc
	systemHealthNumericSensorState *prometheus.Desc
	systemHealthStatusSensor       *prometheus.Desc

	// only used when advancedStorageMetrics == true
	hbaStatus                *prometheus.Desc
	hbaIscsiSendTargetInfo   *prometheus.Desc
	hbaIscsiStaticTargetInfo *prometheus.Desc
	multipathPathState       *prometheus.Desc
	// iscsiDiskInfo            *prometheus.Desc
	scsiLunMounted    *prometheus.Desc
	scsiLunAccessible *prometheus.Desc

	vmNumTotal *prometheus.Desc
}

func NewEsxCollector(scraper *scraper.VCenterScraper, cConf CollectorConfig) *esxCollector {
	labels := []string{"id", "name", "datacenter", "cluster"}
	extraLabels := cConf.HostTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	infoLabels := append(labels, "os_version", "vendor", "model", "asset_tag", "service_tag")
	sysNumLabels := append(labels, "sensor_id", "sensor_name", "sensor_type", "sensor_unit")
	sysStatusLabels := append(labels, "sensor_type", "sensor_name")
	hbaLabels := append(labels, "adapter_name", "src_name", "driver", "model")
	iscsiHbaSendTargetLabels := append(labels, "adapter_name", "src_name", "driver", "model", "target_address")
	iscsiHbaStaticTargetLabels := append(labels, "adapter_name", "src_name", "driver", "model", "target_address", "target_name", "discovery_method")
	multipathLabels := append(labels, "adapter_name", "target_address", "target_name", "disk_name", "datastore")
	vmfsLabels := append(labels, "vendor", "model", "disk_name", "datastore")

	return &esxCollector{
		scraper:              scraper,
		enableStorageMetrics: cConf.HostStorageMetrics,
		extraLabels:          extraLabels,
		//GENERAL
		powerState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "power_state"),
			"esx host powerstate", labels, nil),
		connectionState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "connection_state"),
			"esx host connectionstate", labels, nil),
		maintenance: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "maintenance"),
			"esx host in maintenance", labels, nil),
		uptimeSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "uptime_seconds"),
			"esx host uptime in seconds", labels, nil),
		rebootRequired: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "reboot_required"),
			"esx reboot required", labels, nil),
		vmNumTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "num_vms"),
			"Total number of VM's on the host", labels, nil),
		info: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "info"),
			"Additional information", infoLabels, nil),

		//CPU
		cpuCoresTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "cpu_cores"),
			"esx number of  cores", labels, nil),
		cpuThreadsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "cpu_threads"),
			"esx number of threads", labels, nil),
		availCPUMhz: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "avail_cpu_mhz"),
			"esx total cpu in mhz", labels, nil),
		usedCPUMhz: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "used_cpu_mhz"),
			"esx cpu usage in mhz", labels, nil),

		//MEMORY
		availMemBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "avail_mem_bytes"),
			"esx total memory in bytes", labels, nil),
		usedMemBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "used_mem_bytes"),
			"esx used memory in bytes", labels, nil),

		overallStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "overall_status"),
			"overall health status", labels, nil),
		systemHealthNumericSensorValue: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "system_health_numeric_sensor_value"),
			"Numeric system hardware sensors", sysNumLabels, nil),
		systemHealthNumericSensorState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "system_health_numeric_sensor_state"),
			"Numeric system hardware sensors", sysNumLabels, nil),
		systemHealthStatusSensor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "system_health_status_sensor"),
			"system hardware status sensors", sysStatusLabels, nil),

		//STORAGE
		hbaStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "hba_status"),
			"HBA status", hbaLabels, nil),
		hbaIscsiSendTargetInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "hba_iscsi_send_target_info"),
			"The configured iSCSI send target entries", iscsiHbaSendTargetLabels, nil),
		hbaIscsiStaticTargetInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "hba_iscsi_static_target_info"),
			"The configured iSCSI static target entries.", iscsiHbaStaticTargetLabels, nil),
		multipathPathState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "multipath_path_state"),
			"Multipath path state", multipathLabels, nil),
		scsiLunMounted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "datastore_mounted"),
			"VMFS Datastore mount status", vmfsLabels, nil),
		scsiLunAccessible: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "datastore_accessible"),
			"VMFS Datastore accessible status", vmfsLabels, nil),
	}

}

func (c *esxCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.powerState
	ch <- c.connectionState
	ch <- c.maintenance
	ch <- c.uptimeSeconds
	ch <- c.rebootRequired
	ch <- c.cpuCoresTotal
	ch <- c.cpuThreadsTotal
	ch <- c.availCPUMhz
	ch <- c.usedCPUMhz
	ch <- c.availMemBytes
	ch <- c.usedMemBytes
	ch <- c.overallStatus
	ch <- c.systemHealthNumericSensorState
	ch <- c.systemHealthNumericSensorValue
	ch <- c.systemHealthStatusSensor
	ch <- c.scsiLunMounted
	ch <- c.scsiLunAccessible
	ch <- c.hbaIscsiSendTargetInfo
	ch <- c.hbaIscsiStaticTargetInfo
	ch <- c.hbaStatus
	ch <- c.multipathPathState
	ch <- c.vmNumTotal
	ch <- c.info
}

func (c *esxCollector) Collect(ch chan<- prometheus.Metric) {
	if c.scraper.Host == nil {
		return
	}
	hostData := c.scraper.Host.GetAllSnapshots()
	for _, snap := range hostData {
		timestamp, h := snap.Timestamp, snap.Item

		summary := h.Summary
		qs := summary.QuickStats

		parentChain := c.scraper.GetParentChain(h.Self)

		powerState := ConvertHostSystemPowerStateToValue(summary.Runtime.PowerState)
		connState := ConvertHostSystemConnectionStateToValue(summary.Runtime.ConnectionState)
		maintenance := b2f(h.Runtime.InMaintenanceMode)
		extraLabelValues := func() []string {
			result := []string{}
			for _, tagCat := range c.extraLabels {
				tag := c.scraper.Tags.GetTag(h.Self, tagCat)
				if tag != nil {
					result = append(result, tag.Name)
				} else {
					result = append(result, "")
				}
			}
			return result
		}()
		labelValues := []string{me2id(h.ManagedEntity), h.Name, parentChain.DC, parentChain.Cluster}
		labelValues = append(labelValues, extraLabelValues...)

		os_version := summary.Config.Product.FullName
		vendor := summary.Hardware.Vendor
		model := summary.Hardware.Model
		asset_tag := ""
		service_tag := ""
		for _, i := range summary.Hardware.OtherIdentifyingInfo {
			if i.IdentifierType.GetElementDescription().Key == "AssetTag" {
				asset_tag = i.IdentifierValue
			}
			if i.IdentifierType.GetElementDescription().Key == "ServiceTag" {
				service_tag = i.IdentifierValue
			}
		}
		infoLabelValues := append(labelValues, os_version, vendor, model, asset_tag, service_tag)

		if h.Runtime.HealthSystemRuntime != nil {
			if h.Runtime.HealthSystemRuntime.SystemHealthInfo != nil {
				for _, info := range h.Runtime.HealthSystemRuntime.SystemHealthInfo.NumericSensorInfo {
					sysLabelsValues := append(labelValues, info.Id, info.Name, info.SensorType, info.BaseUnits)
					ch <- prometheus.NewMetricWithTimestamp(timestamp,
						prometheus.MustNewConstMetric(
							c.systemHealthNumericSensorValue, prometheus.GaugeValue, float64(info.CurrentReading), sysLabelsValues...,
						))
					ch <- prometheus.NewMetricWithTimestamp(timestamp,
						prometheus.MustNewConstMetric(
							c.systemHealthNumericSensorState, prometheus.GaugeValue, ConvertHostSystemHardwareStateToValue(info.HealthState.GetElementDescription()), sysLabelsValues...,
						))

				}
			}
			if h.Runtime.HealthSystemRuntime.HardwareStatusInfo != nil {
				for _, info := range h.Runtime.HealthSystemRuntime.HardwareStatusInfo.MemoryStatusInfo {
					elementInfo := info.GetHostHardwareElementInfo()
					sysLabelsValues := append(labelValues, "memory", elementInfo.Name)
					status := ConvertHostSystemHardwareStateToValue(elementInfo.Status.GetElementDescription())
					ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
						c.systemHealthStatusSensor, prometheus.GaugeValue, status, sysLabelsValues...,
					))
				}
				for _, info := range h.Runtime.HealthSystemRuntime.HardwareStatusInfo.CpuStatusInfo {
					elementInfo := info.GetHostHardwareElementInfo()
					sysLabelsValues := append(labelValues, "cpu", elementInfo.Name)
					status := ConvertHostSystemHardwareStateToValue(elementInfo.Status.GetElementDescription())
					ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
						c.systemHealthStatusSensor, prometheus.GaugeValue, status, sysLabelsValues...,
					))
				}
			}
		}

		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.info, prometheus.GaugeValue, 1, infoLabelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.powerState, prometheus.GaugeValue, powerState, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.connectionState, prometheus.GaugeValue, connState, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.maintenance, prometheus.GaugeValue, maintenance, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.uptimeSeconds, prometheus.GaugeValue, float64(summary.QuickStats.Uptime), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.rebootRequired, prometheus.GaugeValue, b2f(summary.RebootRequired), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.cpuCoresTotal, prometheus.GaugeValue, float64(summary.Hardware.NumCpuCores), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.cpuThreadsTotal, prometheus.GaugeValue, float64(summary.Hardware.NumCpuThreads), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.availCPUMhz, prometheus.GaugeValue, float64(int64(summary.Hardware.NumCpuCores)*int64(summary.Hardware.CpuMhz)), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.usedCPUMhz, prometheus.GaugeValue, float64(qs.OverallCpuUsage), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.availMemBytes, prometheus.GaugeValue, float64(summary.Hardware.MemorySize), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.usedMemBytes, prometheus.GaugeValue, float64(int64(qs.OverallMemoryUsage)*int64(1024*1024)), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, ConvertManagedEntityStatusToValue(h.OverallStatus), labelValues...,
		))

		vmsOnHost := c.scraper.VM.GetHostVMs(snap.Item.Self)
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.vmNumTotal, prometheus.GaugeValue, float64(len(vmsOnHost)), labelValues...,
		))
		if c.enableStorageMetrics && h.Config != nil {
			if h.Config.StorageDevice != nil {
				//HBA's
				for _, adapter := range h.Config.StorageDevice.HostBusAdapter {
					hbaInterface := reflect.ValueOf(adapter).Elem().Interface()
					switch hba := hbaInterface.(type) {
					case types.HostInternetScsiHba:
						hbaLabelValues := append(labelValues, hba.Device, hba.IScsiName, hba.Driver, cleanString(hba.Model))
						ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
							c.hbaStatus, prometheus.GaugeValue, ConvertHBAStatusToValue(hba.Status), hbaLabelValues...,
						))
						for _, target := range hba.ConfiguredSendTarget {
							iscsiLabelTargetValues := append(hbaLabelValues, fmt.Sprintf("%s:%d", target.Address, target.Port))
							ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
								c.hbaIscsiSendTargetInfo, prometheus.GaugeValue, 1, iscsiLabelTargetValues...,
							))
						}
						for _, target := range hba.ConfiguredStaticTarget {
							iscsiLabelTargetValues := append(hbaLabelValues, fmt.Sprintf("%s:%d", target.Address, target.Port), target.IScsiName, target.DiscoveryMethod)
							ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
								c.hbaIscsiStaticTargetInfo, prometheus.GaugeValue, 1, iscsiLabelTargetValues...,
							))
						}
					// case types.HostBlockHba:
					// case types.HostFibreChannelHba:
					// case types.HostParallelScsiHba:
					// case types.HostPcieHba:
					// case types.HostRdmaDevice:
					// case types.HostSerialAttachedHba:
					// case types.HostTcpHba:
					default:
						baseAdapter := adapter.GetHostHostBusAdapter()
						hbaLabelValues := append(labelValues, baseAdapter.Device, "", baseAdapter.Driver, cleanString(baseAdapter.Model))
						status := adapter.GetHostHostBusAdapter().Status
						ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
							c.hbaStatus, prometheus.GaugeValue, ConvertHBAStatusToValue(status), hbaLabelValues...,
						))
					}
				}

				//MultiPath path state
				multipaths := getHostMultiPathPaths(h)
				for _, p := range multipaths {
					pathLabelValues := append(labelValues, p.Device, p.TargetAddress, p.TargetIQN, p.NAA, p.DatastoreName)
					ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
						c.multipathPathState, prometheus.GaugeValue, p.Value(), pathLabelValues...,
					))
				}

				//Datastore Mount state
				for _, lun := range getAllScsiLuns(h) {
					vmfsLabelValues := append(labelValues, lun.Vendor, lun.Model, lun.CanonicalName, lun.Datastore)
					ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
						c.scsiLunMounted, prometheus.GaugeValue, b2f(lun.Mounted), vmfsLabelValues...,
					))
					ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
						c.scsiLunAccessible, prometheus.GaugeValue, b2f(lun.Accessable), vmfsLabelValues...,
					))
				}
			}
		}
	}
}

func getHostHBADeviceNameLookupFunc(host mo.HostSystem) func(hbaKey string) string {
	hbaDevices := map[string]string{}
	for _, adapter := range host.Config.StorageDevice.HostBusAdapter {
		hbaDevices[adapter.GetHostHostBusAdapter().Key] = adapter.GetHostHostBusAdapter().Device
	}
	return func(hbaKey string) string {
		if val, ok := hbaDevices[hbaKey]; ok {
			return val
		}
		return ""
	}
}

func getHostLunCanonicalNameLookupFunc(host mo.HostSystem) func(lunID string) string {
	naas := map[string]string{}
	for _, lun := range host.Config.StorageDevice.ScsiLun {
		naas[lun.GetScsiLun().Uuid] = lun.GetScsiLun().CanonicalName
	}

	return func(lunID string) string {
		if val, ok := naas[lunID]; ok {
			return val
		}
		return ""
	}
}

func getHostScsiLunLookupFunc(host mo.HostSystem) func(canonicalName string) scsiLun {
	scsiLuns := map[string]scsiLun{}

	for _, l := range getAllScsiLuns(host) {
		scsiLuns[l.CanonicalName] = l
	}

	return func(canonicalName string) scsiLun {
		if val, ok := scsiLuns[canonicalName]; ok {
			return val
		}
		return scsiLun{}
	}
}

type iscsiPath struct {
	Device        string
	NAA           string
	DatastoreName string
	TargetAddress string
	TargetIQN     string
	State         string
}

func (p *iscsiPath) Value() float64 {
	s := types.MultipathState(p.State)
	if s == types.MultipathStateDead {
		return 1.0
	} else if s == types.MultipathStateDisabled {
		return 2.0
	} else if s == types.MultipathStateStandby {
		return 3.0
	} else if s == types.MultipathStateActive {
		return 4.0
	}
	return 0
}

func getHostMultiPathPaths(host mo.HostSystem) []iscsiPath {
	hbaDeviceNameFunc := getHostHBADeviceNameLookupFunc(host)
	lunCanonnicalNameLookup := getHostLunCanonicalNameLookupFunc(host)
	scsiLunLookupFunc := getHostScsiLunLookupFunc(host)

	iscsiPaths := []iscsiPath{}
	for _, logicalUnit := range host.Config.StorageDevice.MultipathInfo.Lun {
		uuid := logicalUnit.Id
		naa := lunCanonnicalNameLookup(uuid)
		scsiLun := scsiLunLookupFunc(naa)
		for _, path := range logicalUnit.Path {
			device := hbaDeviceNameFunc(path.Adapter)

			transportInterface := reflect.ValueOf(path.Transport).Elem().Interface()
			switch transport := transportInterface.(type) {
			case types.HostInternetScsiTargetTransport:
				for _, address := range transport.Address {
					iscsiPaths = append(iscsiPaths, iscsiPath{
						Device:        device,
						NAA:           naa,
						TargetIQN:     transport.IScsiName,
						TargetAddress: address,
						DatastoreName: scsiLun.Datastore,
						State:         path.State,
					})
				}
			default:
				iscsiPaths = append(iscsiPaths, iscsiPath{
					Device:        device,
					NAA:           naa,
					TargetIQN:     "",
					TargetAddress: "",
					DatastoreName: scsiLun.Datastore,
					State:         path.State,
				})
			}
		}
	}
	return iscsiPaths
}

type scsiLun struct {
	CanonicalName string
	Vendor        string
	Model         string
	Datastore     string
	Accessable    bool
	Mounted       bool
}

func getAllScsiLuns(host mo.HostSystem) []scsiLun {
	mountInfoMap := map[string]types.HostFileSystemMountInfo{}
	if host.Config.FileSystemVolume != nil {
		for _, info := range host.Config.FileSystemVolume.MountInfo {
			volumeInterface := reflect.ValueOf(info.Volume).Elem().Interface()
			switch volume := volumeInterface.(type) {
			case types.HostVmfsVolume:
				for _, extend := range volume.Extent {
					mountInfoMap[extend.DiskName] = info
				}
			default:
				continue
			}
		}
	}

	scsiLuns := map[string]types.ScsiLun{}
	if host.Config.StorageDevice != nil {
		for _, l := range host.Config.StorageDevice.ScsiLun {
			lun := l.GetScsiLun()
			if lun != nil {
				scsiLuns[lun.CanonicalName] = *lun
			}
		}
	}

	result := []scsiLun{}
	if host.Config.StorageDevice != nil {
		for _, l := range host.Config.StorageDevice.ScsiLun {
			lun := l.GetScsiLun()
			if mountInfo, ok := mountInfoMap[lun.CanonicalName]; ok {
				volume := mountInfo.Volume.GetHostFileSystemVolume()
				result = append(result, scsiLun{
					CanonicalName: lun.CanonicalName,
					Model:         cleanString(lun.Model),
					Vendor:        cleanString(lun.Vendor),
					Datastore: func() string {
						if volume != nil {
							return volume.Name
						}
						return "unknown"
					}(),
					Mounted: func() bool {
						if mountInfo.MountInfo.Mounted != nil {
							return *mountInfo.MountInfo.Mounted
						}
						return false
					}(),
					Accessable: func() bool {
						if mountInfo.MountInfo.Accessible != nil {
							return *mountInfo.MountInfo.Accessible
						}
						return false
					}(),
				})
			}

		}
	}
	return result
}

// Converters
func ConvertHostSystemPowerStateToValue(s types.HostSystemPowerState) float64 {
	if s == types.HostSystemPowerStateStandBy {
		return 1.0
	} else if s == types.HostSystemPowerStatePoweredOn {
		return 2.0
	}
	return 0
}

func ConvertHostSystemConnectionStateToValue(s types.HostSystemConnectionState) float64 {
	if s == types.HostSystemConnectionStateNotResponding {
		return 1.0
	} else if s == types.HostSystemConnectionStateConnected {
		return 2.0
	}
	return 0
}

func ConvertHostSystemStandbyModeToValue(s types.HostStandbyMode) float64 {
	if s == types.HostStandbyModeExiting {
		return 1.0
	} else if s == types.HostStandbyModeEntering {
		return 2.0
	} else if s == types.HostStandbyModeIn {
		return 3.0
	}
	return 0
}

func ConvertHostSystemHardwareStateToValue(s *types.ElementDescription) float64 {
	if s == nil {
		return 0
	} else if strings.EqualFold(s.Key, "Red") {
		return 1.0
	} else if strings.EqualFold(s.Key, "Yellow") {
		return 2.0
	} else if strings.EqualFold(s.Key, "Green") {
		return 3.0
	}
	return 0
}

func ConvertMultipathStateToValue(s types.MultipathState) float64 {
	if s == types.MultipathStateDead {
		return 1.0
	} else if s == types.MultipathStateDisabled {
		return 2.0
	} else if s == types.MultipathStateStandby {
		return 3.0
	} else if s == types.MultipathStateActive {
		return 4.0
	}
	return 0
}

func ConvertHBAStatusToValue(status string) float64 {
	// Valid values include "online", "offline", "unbound", and "unknown".
	if strings.EqualFold(status, "unbound") {
		return 1.0
	} else if strings.EqualFold(status, "offline") {
		return 2.0
	} else if strings.EqualFold(status, "online") {
		return 3.0
	}
	return 0
}
