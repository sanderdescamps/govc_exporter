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
	generalCollectorSubsystem = "scraper"
)

type scraperCollector struct {
	scraper                *scraper.VCenterScraper
	sensorRefreshTime      *prometheus.Desc
	sensorRefreshQueryTime *prometheus.Desc
	sensorClientWaitTime   *prometheus.Desc
	sensorRefreshStatus    *prometheus.Desc
	sensorAvailable        *prometheus.Desc
	tcpConnectionCheck     *prometheus.Desc
}

func NewScraperCollector(scraper *scraper.VCenterScraper) *scraperCollector {
	sensorLabels := []string{"sensor"}
	return &scraperCollector{
		scraper: scraper,
		sensorRefreshTime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, generalCollectorSubsystem, "refresh_time"),
			"total time to refresh sensor info in µs", sensorLabels, nil),
		sensorRefreshQueryTime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, generalCollectorSubsystem, "refresh_query_time"),
			"time to query vcenter in µs", sensorLabels, nil),
		sensorClientWaitTime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, generalCollectorSubsystem, "client_wait_time"),
			"time sensor need to wait for a client in µs", sensorLabels, nil),
		sensorRefreshStatus: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, generalCollectorSubsystem, "refresh_status"),
			"refresh status", sensorLabels, nil),
		sensorAvailable: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, generalCollectorSubsystem, "sensor_available"),
			"is sensor enabled", sensorLabels, nil),
		tcpConnectionCheck: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, generalCollectorSubsystem, "tcp_connection_check"),
			"tcp connection check with vcenter", []string{"url", "err"}, nil),
	}
}

func (c *scraperCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.tcpConnectionCheck
	ch <- c.sensorRefreshTime
	ch <- c.sensorRefreshQueryTime
	ch <- c.sensorClientWaitTime
	ch <- c.sensorRefreshStatus
	ch <- c.sensorAvailable
}

func (c *scraperCollector) Collect(ch chan<- prometheus.Metric) {

	status := c.scraper.Status()
	ch <- prometheus.MustNewConstMetric(
		c.tcpConnectionCheck, prometheus.GaugeValue, b2f(status.TCPStatusCheck), status.TCPStatusCheckEndpoint, status.TCPStatusCheckMgs,
	)

	for k, v := range status.SensorAvailable {
		ch <- prometheus.MustNewConstMetric(
			c.sensorAvailable, prometheus.GaugeValue, b2f(v), k,
		)
	}

	for _, m := range status.SensorMetric {
		ch <- prometheus.MustNewConstMetric(
			c.sensorRefreshTime, prometheus.GaugeValue, float64(m.TotalRefreshTime().Microseconds()), m.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.sensorRefreshQueryTime, prometheus.GaugeValue, float64(m.QueryTime.Microseconds()), m.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.sensorClientWaitTime, prometheus.GaugeValue, float64(m.ClientWaitTime.Microseconds()), m.Name,
		)
	}

}
