package collector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

type VMPerfCollector struct {
	scraper     *scraper.VCenterScraper
	extraLabels []string

	perfMetric *prometheus.Desc
}

func NewVMPerfCollector(scraper *scraper.VCenterScraper, cConf config.CollectorConfig) *VMPerfCollector {
	labels := []string{"uuid", "name", "template", "vm_id"}
	extraLabels := cConf.VMTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	perfLabels := append(labels, "kind", "instance", "unit")

	return &VMPerfCollector{
		scraper:     scraper,
		extraLabels: extraLabels,
		perfMetric: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, virtualMachineCollectorSubsystem, "perf_metric"),
			"Performance metric", perfLabels, nil),
	}
}

func (c *VMPerfCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.perfMetric
}

func (c *VMPerfCollector) Collect(ch chan<- prometheus.Metric) {
	if c.scraper.Host == nil || c.scraper.VM == nil || c.scraper.VMPerf == nil {
		fmt.Printf("Can't collect metrics if Host, VM and VMPerf sensor are not defined\n")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), COLLECT_TIMEOUT)
	defer cancel()

	vms, err := c.scraper.DB.GetAllVM(ctx)
	if err != nil && Logger != nil {
		Logger.Error("failed to get vm's", "err", err)
	}
	for _, vm := range vms {
		extraLabelValues := []string{}
		objectTags := c.scraper.DB.GetTags(ctx, vm.Self)
		for _, tagCat := range c.extraLabels {
			extraLabelValues = append(extraLabelValues, objectTags.GetTag(tagCat))
		}

		labelValues := []string{vm.UUID, vm.Name, strconv.FormatBool(vm.Template), vm.Self.ID()}
		labelValues = append(labelValues, extraLabelValues...)

		for metric := range c.scraper.MetricsDB.PopAllVmMetricsIter(ctx, vm.Self) {
			perfMetricLabelValues := append(labelValues, metric.Name, metric.Instance, metric.Unit)
			ch <- prometheus.NewMetricWithTimestamp(metric.Timestamp, prometheus.MustNewConstMetric(
				c.perfMetric, prometheus.GaugeValue, metric.Value, perfMetricLabelValues...,
			))
		}
	}
}
