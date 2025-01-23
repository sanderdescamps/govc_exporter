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
	"github.com/intrinsec/govc_exporter/collector/scraper"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	storagePodCollectorSubsystem = "spod"
)

type storagePodCollector struct {
	scraper       *scraper.VCenterScraper
	extraLabels   []string
	capacity      *prometheus.Desc
	freeSpace     *prometheus.Desc
	overallStatus *prometheus.Desc
}

func NewStoragePodCollector(scraper *scraper.VCenterScraper, cConf CollectorConfig) *storagePodCollector {
	labels := []string{"id", "name", "datacenter"}

	extraLabels := cConf.StoragePodTagLabels
	if len(extraLabels) != 0 {
		labels = append(labels, extraLabels...)
	}

	return &storagePodCollector{
		scraper:     scraper,
		extraLabels: extraLabels,
		capacity: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, storagePodCollectorSubsystem, "capacity_bytes"),
			"storagePod capacity in bytes", labels, nil),
		freeSpace: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, storagePodCollectorSubsystem, "free_space_bytes"),
			"storagePod freespace in bytes", labels, nil),
		overallStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, storagePodCollectorSubsystem, "overall_status"),
			"overall health status", labels, nil),
	}
}

func (c *storagePodCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.capacity
	ch <- c.freeSpace
	ch <- c.overallStatus
}

func (c *storagePodCollector) Collect(ch chan<- prometheus.Metric) {
	if c.scraper.SPOD == nil {
		return
	}

	storagePodData := c.scraper.SPOD.GetAllSnapshots()
	for _, snap := range storagePodData {
		timestamp, s := snap.Timestamp, snap.Item
		summary := s.Summary
		parentChain := c.scraper.GetParentChain(s.Self)

		extraLabelValues := func() []string {
			result := []string{}

			for _, tagCat := range c.extraLabels {
				tag := c.scraper.Tags.GetTag(s.Self, tagCat)
				if tag != nil {
					result = append(result, tag.Name)
				} else {
					result = append(result, "")
				}
			}
			return result
		}()

		labelValues := []string{me2id(s.ManagedEntity), s.Name, parentChain.DC}
		labelValues = append(labelValues, extraLabelValues...)
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.capacity, prometheus.GaugeValue, float64(summary.Capacity), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.freeSpace, prometheus.GaugeValue, float64(summary.FreeSpace), labelValues...,
		))
		ch <- prometheus.NewMetricWithTimestamp(timestamp, prometheus.MustNewConstMetric(
			c.overallStatus, prometheus.GaugeValue, ConvertManagedEntityStatusToValue(s.OverallStatus), labelValues...,
		))
	}
}
