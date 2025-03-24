package collector

import (
	"reflect"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	datastoreCollectorSubsystem = "ds"
)

type datastoreCollector struct {
	// vcCollector
	scraper          *scraper.VCenterScraper
	extraLabels      []string
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

func NewDatastoreCollector(scraper *scraper.VCenterScraper, cConf Config) *datastoreCollector {
	labels := []string{"id", "name", "cluster", "kind"}

	extraLabels := cConf.DatastoreTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	hostLables := append(labels, "esx", "esx_id")
	vmfsLabels := append(labels, "uuid", "naa", "ssh", "local")
	return &datastoreCollector{
		scraper:     scraper,
		extraLabels: extraLabels,
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
	if c.scraper.Datastore == nil {
		return
	}
	datastoreData := c.scraper.Datastore.GetAllSnapshots()
	for _, snap := range datastoreData {
		timestamp, d := snap.Timestamp, snap.Item
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
		}

		parentChain := c.scraper.GetParentChain(d.Self)

		extraLabelValues := func() []string {
			result := []string{}

			for _, tagCat := range c.extraLabels {
				tag := c.scraper.Tags.GetTag(d.Self, tagCat)
				if tag != nil {
					result = append(result, tag.Name)
				} else {
					result = append(result, "")
				}
			}
			return result
		}()

		labelValues := []string{me2id(d.ManagedEntity), summary.Name, parentChain.SPOD, kind}
		labelValues = append(labelValues, extraLabelValues...)

		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.accessible, prometheus.GaugeValue, b2f(summary.Accessible), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.capacity, prometheus.GaugeValue, float64(summary.Capacity), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.freeSpace, prometheus.GaugeValue, float64(summary.FreeSpace), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.maintenance, prometheus.GaugeValue, ConvertDatastoreMaintenanceModeStateToValue(summary.MaintenanceMode), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, ConvertManagedEntityStatusToValue(d.OverallStatus), labelValues...,
		))

		for _, host := range d.Host {
			hostEntity := c.scraper.Host.Get(host.Key)
			if hostEntity != nil {
				hostLabelValues := append(labelValues, hostEntity.Name, hostEntity.Self.Value)
				ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
					c.hostAccessible, prometheus.GaugeValue, b2f(*host.MountInfo.Accessible), hostLabelValues...,
				))
				ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
					c.hostMounted, prometheus.GaugeValue, b2f(*host.MountInfo.Mounted), hostLabelValues...,
				))
				ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
					c.hostVmknicActive, prometheus.GaugeValue, b2f(*host.MountInfo.VmknicActive), hostLabelValues...,
				))
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
				ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
					c.vmfsInfo, prometheus.GaugeValue, 1, vmfsLabelValues...,
				))
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
