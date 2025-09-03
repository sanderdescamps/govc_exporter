package collector

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

const (
	esxCollectorSubsystem = "esx"
)

type esxCollector struct {
	// vcCollector
	enableStorageMetrics bool
	extraLabels          []string

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
	volumeMounted            *prometheus.Desc
	volumeAccessible         *prometheus.Desc
	scsiLunActivePath        *prometheus.Desc
	scsiLunTotalPath         *prometheus.Desc

	vmNumTotal *prometheus.Desc
}

func NewEsxCollector(scraper *scraper.VCenterScraper, cConf config.CollectorConfig) *esxCollector {
	labels := []string{"id", "name", "datacenter", "cluster"}
	extraLabels := cConf.HostTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	infoLabels := append(slices.Clone(labels), "os_version", "vendor", "model", "asset_tag", "service_tag", "bios_version")
	sysNumLabels := append(slices.Clone(labels), "sensor_id", "sensor_name", "sensor_type", "sensor_unit")
	sysStatusLabels := append(slices.Clone(labels), "sensor_type", "sensor_name")

	hbaLabels := append(slices.Clone(labels), "adapter_name", "driver", "model")
	iscsiHbaSendTargetLabels := append(slices.Clone(labels), "adapter_name", "driver", "model", "target_address")
	iscsiHbaStaticTargetLabels := append(slices.Clone(labels), "adapter_name", "driver", "model", "target_address", "target_name", "discovery_method", "initiator_name")
	multipathStatusLabels := append(slices.Clone(labels), "path_name", "adapter_name", "target_address", "target_name", "lun", "canonical_name")
	volumeLabels := append(slices.Clone(labels), "uuid", "canonical_name", "datastore", "local", "ssd")
	scsiLunLabels := append(slices.Clone(labels), "vendor", "model", "canonical_name", "local", "ssd")

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
			"Status of HBA cards", hbaLabels, nil),
		hbaIscsiSendTargetInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "hba_iscsi_send_target_info"),
			"The configured iSCSI send target entries", iscsiHbaSendTargetLabels, nil),
		hbaIscsiStaticTargetInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "hba_iscsi_static_target_info"),
			"The configured iSCSI static target entries.", iscsiHbaStaticTargetLabels, nil),
		multipathPathState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "multipath_path_state"),
			"Multipath path state", multipathStatusLabels, nil),
		volumeMounted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "datastore_mounted"),
			"VMFS Datastore mount status", volumeLabels, nil),
		volumeAccessible: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "datastore_accessible"),
			"VMFS Datastore accessible status", volumeLabels, nil),
		scsiLunActivePath: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "scsilun_active_paths"),
			"Number of active paths to lun", scsiLunLabels, nil),
		scsiLunTotalPath: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "scsilun_total_paths"),
			"Total number of paths to lun", scsiLunLabels, nil),
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
	ch <- c.volumeAccessible
	ch <- c.volumeMounted
	ch <- c.hbaIscsiSendTargetInfo
	ch <- c.hbaIscsiStaticTargetInfo
	ch <- c.hbaStatus
	ch <- c.multipathPathState
	ch <- c.scsiLunActivePath
	ch <- c.scsiLunTotalPath
	ch <- c.vmNumTotal
	ch <- c.info
}

func (c *esxCollector) Collect(ch chan<- prometheus.Metric) {
	if !c.scraper.Host.Enabled() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), COLLECT_TIMEOUT)
	defer cancel()

	hosts, err := c.scraper.DB.GetAllHost(ctx)
	if err != nil && Logger != nil {
		Logger.Error("failed to get hosts", "err", err)
	}
	for _, host := range hosts {
		extraLabelValues := []string{}
		objectTags := c.scraper.DB.GetTags(ctx, host.Self)
		for _, tagCat := range c.extraLabels {
			extraLabelValues = append(extraLabelValues, objectTags.GetTag(tagCat))
		}

		labelValues := []string{host.Self.ID(), host.Name, host.Datacenter, host.Cluster}
		labelValues = append(labelValues, extraLabelValues...)
		infoLabelValues := append(slices.Clone(labelValues), host.OSVersion, host.Vendor, host.Model, host.AssetTag, host.ServiceTag, host.BiosVersion)

		for _, health := range host.SystemHealthNumericSensors {
			sysLabelsValues := append(slices.Clone(labelValues), health.ID, health.Name, health.Type, health.Unit)
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
			sysLabelsValues := append(slices.Clone(labelValues), "memory", elementStatus.Name)
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
			c.availMemBytes, prometheus.GaugeValue, host.AvailMemBytes, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.usedMemBytes, prometheus.GaugeValue, host.UsedMemBytes, labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, host.OverallStatusFloat64(), labelValues...,
		))

		ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
			c.vmNumTotal, prometheus.GaugeValue, host.NumberOfVMs, labelValues...,
		))
		if c.enableStorageMetrics {
			for _, hba := range host.HBA {
				hbaLabelValues := append(slices.Clone(labelValues), hba.Device, hba.Driver, hba.Model)
				ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
					c.hbaStatus, prometheus.GaugeValue, hba.StatusFloat64(), hbaLabelValues...,
				))

				if hba.Type == "iscsi" {
					for _, target := range hba.IscsiDiscoveryTarget {
						iscsiLabelTargetValues := append(slices.Clone(hbaLabelValues), fmt.Sprintf("%s:%d", target.Address, target.Port))
						ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
							c.hbaIscsiSendTargetInfo, prometheus.GaugeValue, 1, iscsiLabelTargetValues...,
						))
					}
					for _, target := range hba.IscsiStaticTarget {
						iscsiLabelTargetValues := append(slices.Clone(hbaLabelValues), fmt.Sprintf("%s:%d", target.Address, target.Port), target.IQN, target.DiscoveryMethod, hba.IscsiInitiatorIQN)
						ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
							c.hbaIscsiStaticTargetInfo, prometheus.GaugeValue, 1, iscsiLabelTargetValues...,
						))
					}
				}
			}

			for _, p := range host.MultipathPathInfo {
				pathLabelValues := append(slices.Clone(labelValues), p.Name, p.Adapter, p.IscsiTargetAddress, p.IscsiTargetIQN, strconv.Itoa(p.LUN), p.CanonicalName)
				ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
					c.multipathPathState, prometheus.GaugeValue, p.StateFloat64(), pathLabelValues...,
				))
			}

			for _, lun := range host.Luns {
				vmfsLabelValues := append(slices.Clone(labelValues), lun.Vendor, lun.Model, lun.CanonicalName, strconv.FormatBool(lun.Local), strconv.FormatBool(lun.SSD))
				ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
					c.scsiLunActivePath, prometheus.GaugeValue, float64(lun.ActiveNumberPaths), vmfsLabelValues...,
				))
				ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
					c.scsiLunTotalPath, prometheus.GaugeValue, float64(lun.TotalNumberPaths), vmfsLabelValues...,
				))
			}

			for _, volume := range host.Volumes {
				volumeLabelValues := append(slices.Clone(labelValues), volume.UUID, volume.DiskName, volume.Name, strconv.FormatBool(volume.Local), strconv.FormatBool(volume.SSD))
				ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
					c.volumeAccessible, prometheus.GaugeValue, b2f(volume.Accessible), volumeLabelValues...,
				))
				ch <- prometheus.NewMetricWithTimestamp(host.Timestamp, prometheus.MustNewConstMetric(
					c.volumeMounted, prometheus.GaugeValue, b2f(volume.Mounted), volumeLabelValues...,
				))
			}
		}
	}
}
