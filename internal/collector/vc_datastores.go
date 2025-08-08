package collector

import (
	"context"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
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
	vmfsLabels := append(labels, "uuid", "naa", "ssd", "local")
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
	if !c.scraper.Datastore.Enabled() {
		return
	}
	ctx := context.Background()
	for datastore := range c.scraper.DB.GetAllDatastoreIter(ctx) {

		extraLabelValues := []string{}
		objectTags := c.scraper.DB.GetTags(ctx, datastore.Self)
		for _, tagCat := range c.extraLabels {
			extraLabelValues = append(extraLabelValues, objectTags.GetTag(tagCat))
		}

		labelValues := []string{datastore.Self.ID(), datastore.Name, datastore.DatastoreCluster, datastore.Kind}
		labelValues = append(labelValues, extraLabelValues...)

		ch <- prometheus.NewMetricWithTimestamp(datastore.Timestamp, prometheus.MustNewConstMetric(
			c.accessible, prometheus.GaugeValue, b2f(datastore.Accessible), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(datastore.Timestamp, prometheus.MustNewConstMetric(
			c.capacity, prometheus.GaugeValue, float64(datastore.Capacity), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(datastore.Timestamp, prometheus.MustNewConstMetric(
			c.freeSpace, prometheus.GaugeValue, float64(datastore.FreeSpace), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(datastore.Timestamp, prometheus.MustNewConstMetric(
			c.maintenance, prometheus.GaugeValue, datastore.MaintenanceStatusFloat64(), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(datastore.Timestamp, prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, datastore.OverallStatusFloat64(), labelValues...,
		))

		for _, mountInfo := range datastore.HostMountInfo {
			hostLabelValues := append(labelValues, mountInfo.Host, mountInfo.HostID)
			ch <- prometheus.NewMetricWithTimestamp(datastore.Timestamp, prometheus.MustNewConstMetric(
				c.hostAccessible, prometheus.GaugeValue, b2f(mountInfo.Accessible), hostLabelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(datastore.Timestamp, prometheus.MustNewConstMetric(
				c.hostMounted, prometheus.GaugeValue, b2f(mountInfo.Mounted), hostLabelValues...,
			))
			ch <- prometheus.NewMetricWithTimestamp(datastore.Timestamp, prometheus.MustNewConstMetric(
				c.hostVmknicActive, prometheus.GaugeValue, b2f(mountInfo.VmknicActiveNic), hostLabelValues...,
			))
		}

		if vmfsInfo := datastore.VmfsInfo; vmfsInfo != nil {
			vmfsLabelValues := append(
				labelValues,
				vmfsInfo.UUID,
				vmfsInfo.NAA,
				strconv.FormatBool(vmfsInfo.SSD),
				strconv.FormatBool(vmfsInfo.Local),
			)
			ch <- prometheus.NewMetricWithTimestamp(datastore.Timestamp, prometheus.MustNewConstMetric(
				c.vmfsInfo, prometheus.GaugeValue, 1, vmfsLabelValues...,
			))
		}
	}

}
