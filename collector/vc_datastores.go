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

	"github.com/intrinsec/govc_exporter/collector/scraper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	datastoreCollectorSubsystem = "ds"
)

type datastoreCollector struct {
	// vcCollector
	scraper       *scraper.VCenterScraper
	capacity      *prometheus.Desc
	freeSpace     *prometheus.Desc
	accessible    *prometheus.Desc
	maintenance   *prometheus.Desc
	overallStatus *prometheus.Desc
}

func NewDatastoreCollector(scraper *scraper.VCenterScraper) *datastoreCollector {
	labels := []string{"id", "name", "cluster", "kind"}
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
	}
}

func (c *datastoreCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.accessible
	ch <- c.capacity
	ch <- c.freeSpace
	ch <- c.maintenance
	ch <- c.overallStatus
}

func (c *datastoreCollector) Collect(ch chan<- prometheus.Metric) {
	datastores := c.scraper.Datastore.GetAll()
	for _, d := range datastores {
		summary := d.Summary

		kind := "NONE"

		if d.Info != nil {
			// func() {
			// 	b, err := json.Marshal(info)
			// 	if err != nil {
			// 		return
			// 	}
			// 	var vmfsInfo types.VmfsDatastoreInfo
			// 	err = json.Unmarshal(b, &vmfsInfo)
			// 	if err != nil {
			// 		return
			// 	}
			// 	if vmfsInfo.Vmfs != nil {
			// 		isLocal = strconv.FormatBool(*vmfsInfo.Vmfs.Local)
			// 		isSSD = strconv.FormatBool(*vmfsInfo.Vmfs.Ssd)
			// 	}
			// }()

			iInfo := reflect.ValueOf(d.Info).Elem().Interface()
			switch pInfo := iInfo.(type) {
			case types.LocalDatastoreInfo:
				kind = "local"
			case types.VmfsDatastoreInfo:
				if pInfo.Vmfs != nil {
					var ssd string
					if *pInfo.Vmfs.Ssd {
						ssd = "ssd"
					} else {
						ssd = "hdd"
					}
					var local string
					if pInfo.Vmfs.Local == nil || *pInfo.Vmfs.Local {
						local = "local"
					} else {
						local = "non-local"
					}
					kind = fmt.Sprintf("vmfs-%s-%s", local, ssd)
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
