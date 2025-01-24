// Copyright 2020 Intrinsec
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !noesx
// +build !noesx

package collector

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/intrinsec/govc_exporter/collector/scraper"
	"github.com/prometheus/client_golang/prometheus"
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
	availCPUMhz                    *prometheus.Desc
	usedCPUMhz                     *prometheus.Desc
	availMemBytes                  *prometheus.Desc
	usedMemBytes                   *prometheus.Desc
	overallStatus                  *prometheus.Desc
	systemHealthNumericSensorValue *prometheus.Desc
	systemHealthNumericSensorState *prometheus.Desc
	systemHealthStatusSensor       *prometheus.Desc

	// only used when advancedStorageMetrics == true
	hbaStatus                *prometheus.Desc
	hbaIscsiSendTargetInfo   *prometheus.Desc
	hbaIscsiStaticTargetInfo *prometheus.Desc
	multipathPathState       *prometheus.Desc
	iscsiDiskInfo            *prometheus.Desc
}

func NewEsxCollector(scraper *scraper.VCenterScraper, cConf CollectorConfig) *esxCollector {
	labels := []string{"id", "name", "datacenter", "cluster"}
	extraLabels := cConf.HostTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	sysNumLabels := append(labels, "sensor_id", "sensor_name", "sensor_type", "sensor_unit")
	sysStatusLabels := append(labels, "sensor_type", "sensor_name")
	hbaLabels := append(labels, "adapter_name", "src_name", "driver", "model")
	iscsiHbaSendTargetLabels := append(hbaLabels, "target_address")
	iscsiHbaStaticTargetLabels := append(hbaLabels, "target_address", "target_name", "discovery_method")
	multipathLabels := append(labels, "path_name", "adapter_name", "target_address", "target_name", "disk_name", "datastore")
	iscsiLabels := append(labels, "vendor", "model", "disk_name", "ssd", "local", "datastore")

	return &esxCollector{
		scraper:              scraper,
		enableStorageMetrics: cConf.HostStorageMetrics,
		extraLabels:          extraLabels,
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
		cpuCoresTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "cpu_cores_total"),
			"esx number of  cores", labels, nil),
		availCPUMhz: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "avail_cpu_mhz"),
			"esx total cpu in mhz", labels, nil),
		usedCPUMhz: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "used_cpu_mhz"),
			"esx cpu usage in mhz", labels, nil),
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
		iscsiDiskInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "iscsi_disk_info"),
			"Multipath path state", iscsiLabels, nil),
	}

}

func (c *esxCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.powerState
	ch <- c.connectionState
	ch <- c.maintenance
	ch <- c.uptimeSeconds
	ch <- c.rebootRequired
	ch <- c.cpuCoresTotal
	ch <- c.availCPUMhz
	ch <- c.usedCPUMhz
	ch <- c.availMemBytes
	ch <- c.usedMemBytes
	ch <- c.overallStatus
	ch <- c.systemHealthNumericSensorState
	ch <- c.systemHealthNumericSensorValue
	ch <- c.systemHealthStatusSensor
	ch <- c.iscsiDiskInfo
	ch <- c.hbaIscsiSendTargetInfo
	ch <- c.hbaIscsiStaticTargetInfo
	ch <- c.hbaStatus
	ch <- c.multipathPathState
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

		if h.Runtime.HealthSystemRuntime != nil {
			if h.Runtime.HealthSystemRuntime.SystemHealthInfo != nil {
				for _, info := range h.Runtime.HealthSystemRuntime.SystemHealthInfo.NumericSensorInfo {
					sysLabelsValues := append(labelValues, info.Id, info.Name, info.SensorType, info.BaseUnits)
					timestamp, err := time.Parse(time.RFC3339Nano, info.TimeStamp)
					if err != nil {
						timestamp = time.Now()
					}
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

		if c.enableStorageMetrics && h.Config != nil {
			hbaDeviceLookup := map[string]string{}
			if h.Config.StorageDevice != nil {
				for _, adapter := range h.Config.StorageDevice.HostBusAdapter {
					hbaDeviceLookup[adapter.GetHostHostBusAdapter().Key] = adapter.GetHostHostBusAdapter().Device

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

				for _, logicalUnit := range h.Config.StorageDevice.MultipathInfo.Lun {
					for _, path := range logicalUnit.Path {
						device := ""
						if d, ok := hbaDeviceLookup[path.Adapter]; ok {
							device = d
						}

						uuid := logicalUnit.Id
						naa := func() string {
							for _, lun := range h.Config.StorageDevice.ScsiLun {
								if lun.GetScsiLun().Uuid == uuid {
									return lun.GetScsiLun().CanonicalName
								}
							}
							return ""
						}()

						datastoreName := func() string {
							if h.Config.FileSystemVolume != nil {
								for _, mountInfo := range h.Config.FileSystemVolume.MountInfo {
									volumeInterface := reflect.ValueOf(mountInfo.Volume).Elem().Interface()
									switch volume := volumeInterface.(type) {
									case types.HostVmfsVolume:
										for _, extend := range volume.Extent {
											if extend.DiskName == naa {
												return volume.Name
											}
										}
									default:
										continue
									}

								}
							}
							return "unknown"
						}()

						transportInterface := reflect.ValueOf(path.Transport).Elem().Interface()
						switch transport := transportInterface.(type) {
						case types.HostInternetScsiTargetTransport:
							for _, address := range transport.Address {
								pathLabelValues := append(labelValues, path.Name, device, address, transport.IScsiName, naa, datastoreName)
								ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
									c.multipathPathState, prometheus.GaugeValue, ConvertMultipathStateToValue(types.MultipathState(path.State)), pathLabelValues...,
								))
							}
						default:
							pathLabelValues := append(labelValues, path.Name, device, "", "", naa, datastoreName)
							ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
								c.multipathPathState, prometheus.GaugeValue, ConvertMultipathStateToValue(types.MultipathState(path.State)), pathLabelValues...,
							))
						}
					}

				}

				for _, baseScsiLun := range h.Config.StorageDevice.ScsiLun {
					canonicalName := baseScsiLun.GetScsiLun().CanonicalName
					datastoreName := func() string {
						if h.Config.FileSystemVolume != nil {
							for _, mountInfo := range h.Config.FileSystemVolume.MountInfo {
								volumeInterface := reflect.ValueOf(mountInfo.Volume).Elem().Interface()
								switch volume := volumeInterface.(type) {
								case types.HostVmfsVolume:
									for _, extend := range volume.Extent {
										if extend.DiskName == canonicalName {
											return volume.Name
										}
									}
								default:
									continue
								}

							}
						}
						return "unknown"
					}()

					scsiLunInterface := reflect.ValueOf(baseScsiLun).Elem().Interface()
					switch scsiLun := scsiLunInterface.(type) {
					case types.HostScsiDisk:
						iscsiLabelValues := append(labelValues, cleanString(scsiLun.Vendor), cleanString(scsiLun.Model), cleanString(scsiLun.CanonicalName), strconv.FormatBool(*scsiLun.Ssd), strconv.FormatBool(*scsiLun.LocalDisk), datastoreName)
						ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
							c.iscsiDiskInfo, prometheus.GaugeValue, 1, iscsiLabelValues...,
						))
					default:
						lun := baseScsiLun.GetScsiLun()
						scsiLabelValues := append(labelValues, cleanString(lun.Vendor), cleanString(lun.Model), cleanString(lun.CanonicalName), "", "", datastoreName)
						ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
							c.iscsiDiskInfo, prometheus.GaugeValue, 1, scsiLabelValues...,
						))
					}
				}
			}
		}
	}
}

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
