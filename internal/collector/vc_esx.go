package collector

import (
	"context"
	"fmt"
	"reflect"

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
	genericHBAStatus *prometheus.Desc
	iscsiHBAStatus   *prometheus.Desc

	hbaIscsiSendTargetInfo   *prometheus.Desc
	hbaIscsiStaticTargetInfo *prometheus.Desc
	multipathPathState       *prometheus.Desc
	// iscsiDiskInfo            *prometheus.Desc
	scsiLunMounted    *prometheus.Desc
	scsiLunAccessible *prometheus.Desc

	vmNumTotal *prometheus.Desc
}

func NewEsxCollector(scraper *scraper.VCenterScraper, cConf Config) *esxCollector {
	labels := []string{"id", "name", "datacenter", "cluster"}
	extraLabels := cConf.HostTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	infoLabels := append(labels, "os_version", "vendor", "model", "asset_tag", "service_tag", "bios_version")
	sysNumLabels := append(labels, "sensor_id", "sensor_name", "sensor_type", "sensor_unit")
	sysStatusLabels := append(labels, "sensor_type", "sensor_name")
	genericHBALabels := append(labels, "adapter_name", "driver", "model")
	iscsiHBALabels := append(labels, "adapter_name", "iqn", "driver", "model")
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
		genericHBAStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "generic_hba_status"),
			"Status of generic HBA cards", genericHBALabels, nil),
		iscsiHBAStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "iscsi_hba_status"),
			"Status of iscsi HBA card", iscsiHBALabels, nil),
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
	ch <- c.genericHBAStatus
	ch <- c.iscsiHBAStatus
	ch <- c.multipathPathState
	ch <- c.vmNumTotal
	ch <- c.info
}

func (c *esxCollector) Collect(ch chan<- prometheus.Metric) {
	if !c.scraper.Host.Enabled() {
		return
	}
	ctx := context.Background()
	for host := range c.scraper.DB.GetAllHostIter(ctx) {
		extraLabelValues := []string{}
		objectTags := c.scraper.DB.GetTags(ctx, host.Self)
		for _, tagCat := range c.extraLabels {
			extraLabelValues = append(extraLabelValues, objectTags.GetTag(tagCat))
		}

		labelValues := []string{host.Self.ID(), host.Name, host.Datacenter, host.Cluster}
		labelValues = append(labelValues, extraLabelValues...)
		infoLabelValues := append(labelValues, host.OSVersion, host.Vendor, host.Model, host.AssetTag, host.ServiceTag, host.BiosVersion)

		for _, health := range host.SystemHealthNumericSensors {
			sysLabelsValues := append(labelValues, health.ID, health.Name, health.Type, health.Unit)
			ch <- prometheus.NewMetricWithTimestamp(host.Timestamp,
				prometheus.MustNewConstMetric(
					c.systemHealthNumericSensorValue, prometheus.GaugeValue, health.Value, sysLabelsValues...,
				))
			ch <- prometheus.NewMetricWithTimestamp(host.Timestamp,
				prometheus.MustNewConstMetric(
					c.systemHealthNumericSensorState, prometheus.GaugeValue, health.HealthStatus(), sysLabelsValues...,
				))
		}

		for _, elementStatus := range host.HardwareStatus {
			sysLabelsValues := append(labelValues, "memory", elementStatus.Name)
			ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
				c.systemHealthStatusSensor, prometheus.GaugeValue, elementStatus.HealthStatus(), sysLabelsValues...,
			))
		}

		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.info, prometheus.GaugeValue, 1, infoLabelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.powerState, prometheus.GaugeValue, host.PowerStateFloat64(), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.connectionState, prometheus.GaugeValue, host.ConnectionStateFloat64(), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.maintenance, prometheus.GaugeValue, b2f(host.Maintenance), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.uptimeSeconds, prometheus.GaugeValue, host.UptimeSeconds, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.rebootRequired, prometheus.GaugeValue, b2f(host.RebootRequired), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.cpuCoresTotal, prometheus.GaugeValue, host.CPUCoresTotal, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.cpuThreadsTotal, prometheus.GaugeValue, host.CPUThreadsTotal, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.availCPUMhz, prometheus.GaugeValue, host.AvailCPUMhz, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.usedCPUMhz, prometheus.GaugeValue, host.UsedCPUMhz, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.availMemBytes, prometheus.GaugeValue, host.AvailCPUMhz, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.usedMemBytes, prometheus.GaugeValue, host.UsedCPUMhz, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, host.OverallStatusFloat64(), labelValues...,
		))

		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.vmNumTotal, prometheus.GaugeValue, host.NumberOfVMs, labelValues...,
		))
		if c.enableStorageMetrics {
			for _, hba := range host.GenericHBA {
				hbaLabelValues := append(labelValues, hba.Name, hba.Driver, hba.Model)
				ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
					c.genericHBAStatus, prometheus.GaugeValue, hba.StatusFloat64(), hbaLabelValues...,
				))
			}

			for _, hba := range host.IscsiHBA {
				hbaLabelValues := append(labelValues, hba.Name, hba.IQN, hba.Driver, hba.Model)
				ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
					c.iscsiHBAStatus, prometheus.GaugeValue, hba.StatusFloat64(), hbaLabelValues...,
				))
				for _, target := range hba.DiscoveryTarget {
					iscsiLabelTargetValues := append(hbaLabelValues, fmt.Sprintf("%s:%d", target.Address, target.Port))
					ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
						c.hbaIscsiSendTargetInfo, prometheus.GaugeValue, 1, iscsiLabelTargetValues...,
					))
				}
				for _, target := range hba.StaticTarget {
					iscsiLabelTargetValues := append(hbaLabelValues, fmt.Sprintf("%s:%d", target.Address, target.Port), target.IQN, target.DiscoveryMethod)
					ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
						c.hbaIscsiStaticTargetInfo, prometheus.GaugeValue, 1, iscsiLabelTargetValues...,
					))
				}

			}

			for _, p := range host.IscsiMultiPathPaths {
				pathLabelValues := append(labelValues, p.Device, p.TargetAddress, p.TargetIQN, p.NAA, p.DatastoreName)
				ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
					c.multipathPathState, prometheus.GaugeValue, p.StateFloat64(), pathLabelValues...,
				))
			}

			for _, lun := range host.SCSILuns {
				vmfsLabelValues := append(labelValues, lun.Vendor, lun.Model, lun.CanonicalName, lun.Datastore)
				ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
					c.scsiLunMounted, prometheus.GaugeValue, b2f(lun.Mounted), vmfsLabelValues...,
				))
				ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
					c.scsiLunAccessible, prometheus.GaugeValue, b2f(lun.Accessible), vmfsLabelValues...,
				))
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
