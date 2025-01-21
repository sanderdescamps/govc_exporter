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

	"github.com/intrinsec/govc_exporter/collector/scraper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	datastoreCollectorSubsystem = "ds"
)

type datastoreCollector struct {
	// vcCollector
	scraper          *scraper.VCenterScraper
	capacity         *prometheus.Desc
	freeSpace        *prometheus.Desc
	accessible       *prometheus.Desc
	maintenance      *prometheus.Desc
	overallStatus    *prometheus.Desc
	hostAccessible   *prometheus.Desc
	hostMounted      *prometheus.Desc
	hostVmknicActive *prometheus.Desc
	vmfsInfo         *prometheus.Desc
}

func NewDatastoreCollector(scraper *scraper.VCenterScraper) *datastoreCollector {
	labels := []string{"id", "name", "cluster", "kind"}
	hostLables := append(labels, "esx", "esx_id")
	vmfsLabels := append(labels, "uuid", "naa", "ssh", "local")
	return &datastoreCollector{
		scraper: scraper,
		accessible: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, datastoreCollectorSubsystem, "accessible"),
			"datastore is accessible", labels, nil),
		freeSpace: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, datastoreCollectorSubsystem, "free_space_bytes"),
			"datastore freespace in bytes", labels, nil),
		capacity: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, datastoreCollectorSubsystem, "total_capacity_bytes"),
			"datastore capacity in bytes", labels, nil),
		maintenance: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, datastoreCollectorSubsystem, "maintenance"),
			"datastore in maintenance", labels, nil),
		overallStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, datastoreCollectorSubsystem, "overall_status"),
			"overall health status", labels, nil),
		hostAccessible: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, datastoreCollectorSubsystem, "host_accessible"),
			"if datastore is accessible for host", hostLables, nil),
		hostMounted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, datastoreCollectorSubsystem, "host_mounted"),
			"if datastore is mounted to host", hostLables, nil),
		hostVmknicActive: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, datastoreCollectorSubsystem, "host_vmk_nic_active"),
			"Indicates whether vmknic is active or inactive", hostLables, nil),
		vmfsInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, datastoreCollectorSubsystem, "vmfs_info"),
			"Info in case datastore is of type vmsf", vmfsLabels, nil),
	}
}

func (c *datastoreCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.accessible
	ch <- c.capacity
	ch <- c.freeSpace
	ch <- c.maintenance
	ch <- c.overallStatus
	ch <- c.hostAccessible
	ch <- c.hostMounted
	ch <- c.hostVmknicActive
	ch <- c.vmfsInfo
}

func (c *datastoreCollector) Collect(ch chan<- prometheus.Metric) {
	datastores := c.scraper.Datastore.GetAll()
	for _, d := range datastores {
		summary := d.Summary

		kind := "NONE"
		var vmfsInfo *types.HostVmfsVolume = nil

		if d.Info != nil {
			iInfo := reflect.ValueOf(d.Info).Elem().Interface()
			switch parsedInfo := iInfo.(type) {
			case types.LocalDatastoreInfo:
				kind = "local"
			case types.VmfsDatastoreInfo:
				kind = "vmfs"
				if parsedInfo.Vmfs != nil {
					vmfsInfo = parsedInfo.Vmfs
				}
			case types.NasDatastoreInfo:
				kind = "nas"
			case types.PMemDatastoreInfo:
				kind = "pmem"
			case types.VsanDatastoreInfo:
				kind = "vsan"
			case types.VvolDatastoreInfo:
				kind = "vvol"
			}
		} else {
			fmt.Printf("info is nil")
		}

		parentChain := c.scraper.GetParentChain(d.Self)
		labelValues := []string{me2id(d.ManagedEntity), summary.Name, parentChain.SPOD, kind}

		ch <- prometheus.MustNewConstMetric(
			c.accessible, prometheus.GaugeValue, b2f(summary.Accessible), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.capacity, prometheus.GaugeValue, float64(summary.Capacity), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.freeSpace, prometheus.GaugeValue, float64(summary.FreeSpace), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.maintenance, prometheus.GaugeValue, ConvertDatastoreMaintenanceModeStateToValue(summary.MaintenanceMode), labelValues...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, ConvertManagedEntityStatusToValue(d.OverallStatus), labelValues...,
		)

		for _, host := range d.Host {
			hostEntity := c.scraper.Host.Get(host.Key)
			if hostEntity != nil {
				hostLabelValues := append(labelValues, hostEntity.Name, hostEntity.Self.Value)
				ch <- prometheus.MustNewConstMetric(
					c.hostAccessible, prometheus.GaugeValue, b2f(*host.MountInfo.Accessible), hostLabelValues...,
				)
				ch <- prometheus.MustNewConstMetric(
					c.hostMounted, prometheus.GaugeValue, b2f(*host.MountInfo.Mounted), hostLabelValues...,
				)
				ch <- prometheus.MustNewConstMetric(
					c.hostVmknicActive, prometheus.GaugeValue, b2f(*host.MountInfo.VmknicActive), hostLabelValues...,
				)
			}

		}

		if kind == "vmfs" {
			if vmfsInfo != nil {
				vmfsLabelValues := append(
					labelValues,
					vmfsInfo.Uuid,
					func() string {
						for _, extent := range vmfsInfo.Extent {
							return extent.DiskName
						}
						return ""
					}(),
					strconv.FormatBool(*vmfsInfo.Ssd),
					strconv.FormatBool(*vmfsInfo.Local),
				)
				ch <- prometheus.MustNewConstMetric(
					c.vmfsInfo, prometheus.GaugeValue, 1, vmfsLabelValues...,
				)
			}
		}
	}

}

func ConvertDatastoreMaintenanceModeStateToValue(d string) float64 {
	dTyped := types.DatastoreSummaryMaintenanceModeState(d)
	if dTyped == types.DatastoreSummaryMaintenanceModeStateEnteringMaintenance {
		return 1.0
	} else if dTyped == types.DatastoreSummaryMaintenanceModeStateInMaintenance {
		return 2.0
	}
	return 0
}
