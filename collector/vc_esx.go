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
	scraper                   *scraper.VCenterScraper
	powerState                *prometheus.Desc
	connectionState           *prometheus.Desc
	maintenance               *prometheus.Desc
	uptimeSeconds             *prometheus.Desc
	rebootRequired            *prometheus.Desc
	cpuCoresTotal             *prometheus.Desc
	availCPUMhz               *prometheus.Desc
	usedCPUMhz                *prometheus.Desc
	availMemBytes             *prometheus.Desc
	usedMemBytes              *prometheus.Desc
	overallStatus             *prometheus.Desc
	systemHealthNumericSensor *prometheus.Desc
	systemHealthStatusSensor  *prometheus.Desc
}

func NewEsxCollector(scraper *scraper.VCenterScraper) *esxCollector {
	labels := []string{"id", "name", "datacenter", "cluster"}
	sysNumLabels := append(labels, "sensor_id", "sensor_name", "sensor_type", "sensor_unit")
	sysStatusLabels := append(labels, "sensor_name")
	return &esxCollector{
		scraper: scraper,
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
		systemHealthNumericSensor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "system_health_numeric_sensor"),
			"Numeric system hardware sensors", sysNumLabels, nil),
		systemHealthStatusSensor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "system_health_status_sensor"),
			"system hardware status sensors", sysStatusLabels, nil),
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
	ch <- c.systemHealthNumericSensor
	ch <- c.systemHealthStatusSensor
}

func (c *esxCollector) Collect(ch chan<- prometheus.Metric) {
	hosts := c.scraper.Host.GetAll()
	for _, h := range hosts {
		summary := h.Summary
		qs := summary.QuickStats

		parentChain := c.scraper.GetParentChain(h.Self)

		powerState := ConvertHostSystemPowerStateToValue(summary.Runtime.PowerState)
		connState := ConvertHostSystemConnectionStateToValue(summary.Runtime.ConnectionState)
		maintenance := b2f(h.Runtime.InMaintenanceMode)
		labelValues := []string{me2id(h.ManagedEntity), h.Name, parentChain.DC, parentChain.Cluster}

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
							c.systemHealthNumericSensor, prometheus.GaugeValue, float64(info.CurrentReading), sysLabelsValues...,
						))
				}
			}
			if h.Runtime.HealthSystemRuntime.HardwareStatusInfo != nil {
				systemSensors := append(
					h.Runtime.HealthSystemRuntime.HardwareStatusInfo.MemoryStatusInfo,
					h.Runtime.HealthSystemRuntime.HardwareStatusInfo.CpuStatusInfo...,
				)
				for _, info := range systemSensors {
					elementInfo := info.GetHostHardwareElementInfo()
					sysLabelsValues := append(labelValues, elementInfo.Name)
					status := ConvertHostSystemHardwareElementDescriptionToValue(*elementInfo.Status.GetElementDescription())
					ch <- prometheus.MustNewConstMetric(
						c.systemHealthStatusSensor, prometheus.GaugeValue, status, sysLabelsValues...,
					)
				}
			}
		}

		ch <- prometheus.MustNewConstMetric(
			c.powerState, prometheus.GaugeValue, powerState, labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.connectionState, prometheus.GaugeValue, connState, labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.maintenance, prometheus.GaugeValue, maintenance, labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.uptimeSeconds, prometheus.GaugeValue, float64(summary.QuickStats.Uptime), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.rebootRequired, prometheus.GaugeValue, b2f(summary.RebootRequired), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.cpuCoresTotal, prometheus.GaugeValue, float64(summary.Hardware.NumCpuCores), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.availCPUMhz, prometheus.GaugeValue, float64(int64(summary.Hardware.NumCpuCores)*int64(summary.Hardware.CpuMhz)), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.usedCPUMhz, prometheus.GaugeValue, float64(qs.OverallCpuUsage), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.availMemBytes, prometheus.GaugeValue, float64(summary.Hardware.MemorySize), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.usedMemBytes, prometheus.GaugeValue, float64(int64(qs.OverallMemoryUsage)*int64(1024*1024)), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, ConvertManagedEntityStatusToValue(h.OverallStatus), labelValues...,
		)
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

func ConvertHostSystemHardwareElementDescriptionToValue(s types.ElementDescription) float64 {
	if strings.EqualFold(s.Key, "Red") {
		return 1.0
	} else if strings.EqualFold(s.Key, "Yellow") {
		return 2.0
	} else if strings.EqualFold(s.Key, "Green") {
		return 3.0
	}
	return 0
}
