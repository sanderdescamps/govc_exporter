package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/collector/scraper"
)

type VMPerfCollector struct {
	extraLabels []string
	scraper     *scraper.VCenterScraper

	perfMetric *prometheus.Desc
}

func NewVMPerfCollector(scraper *scraper.VCenterScraper, cConf CollectorConfig) *VMPerfCollector {
	labels := []string{"uuid", "name", "template", "vm_id"}
	extraLabels := cConf.VMTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	perfLabels := append(labels, "kind", "unit")

	return &VMPerfCollector{
		scraper:     scraper,
		extraLabels: extraLabels,
		perfMetric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, esxCollectorSubsystem, "perf_metric"),
			"Performance metric", perfLabels, nil),
	}
}

func (c *VMPerfCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.perfMetric
}

func (c *VMPerfCollector) Collect(ch chan<- prometheus.Metric) {
	if c.scraper.Host == nil || c.scraper.HostPerf == nil || c.scraper.VM == nil || c.scraper.VMPerf == nil {
		return
	}

	for _, ref := range c.scraper.VM.GetAllRefs() {
		vm := c.scraper.VM.Get(ref)

		esxName := "NONE"
		hostRef := vm.Runtime.Host
		if hostRef != nil {
			host := c.scraper.Host.Get(*hostRef)
			if host != nil {
				esxName = host.Name
			}
		}

		parentChain := c.scraper.GetParentChain(vm.Self)

		extraLabelValues := func() []string {
			result := []string{}

			for _, tagCat := range c.extraLabels {
				tag := c.scraper.Tags.GetTag(vm.Self, tagCat)
				if tag != nil {
					result = append(result, tag.Name)
				} else {
					result = append(result, "")
				}
			}
			return result
		}()

		labelValues := []string{vm.Config.Uuid, vm.Name, strconv.FormatBool(vm.Config.Template), vm.Self.Value}
		labelValues = append(labelValues, extraLabelValues...)
		labelValues = append(labelValues, parentChain.ResourcePool, parentChain.DC, parentChain.Cluster, esxName)

		for _, metric := range c.scraper.VMPerf.PopAll(ref) {
			perfMetricLabelValues := append(labelValues, metric.Name, metric.Unit)
			ch <- prometheus.NewMetricWithTimestamp(metric.TimeStamp, prometheus.MustNewConstMetric(
				c.perfMetric, prometheus.GaugeValue, metric.Value, perfMetricLabelValues...,
			))
		}
	}
}
