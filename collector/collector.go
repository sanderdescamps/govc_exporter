// Copyright 2015 The Prometheus Authors
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

// Package collector includes all individual collectors to gather and export system metrics.
package collector

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/intrinsec/govc_exporter/collector/helper"
	"github.com/intrinsec/govc_exporter/collector/scraper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Namespace defines the common namespace to be used by all metrics.
const namespace = "govc"

type CollectorConfig struct {
	UseIsecSpecifics       bool
	DisableExporterMetrics bool

	MaxRequests int

	ClusterTagLabels      []string
	DatastoreTagLabels    []string
	HostTagLabels         []string
	ResourcePoolTagLabels []string
	StoragePodTagLabels   []string

	VMAdvancedNetworkMetrics bool
	VMAdvancedStorageMetrics bool
	VMTagLabels              []string

	HostStorageMetrics bool
}

func DefaultCollectorConf() *CollectorConfig {
	return &CollectorConfig{
		UseIsecSpecifics:       false,
		DisableExporterMetrics: false,

		MaxRequests: 10,

		ClusterTagLabels:      []string{},
		DatastoreTagLabels:    []string{},
		HostTagLabels:         []string{},
		ResourcePoolTagLabels: []string{},
		StoragePodTagLabels:   []string{},

		VMAdvancedNetworkMetrics: false,
		VMAdvancedStorageMetrics: false,
		VMTagLabels:              []string{},

		HostStorageMetrics: false,
	}
}

type VCCollector struct {
	scraper    *scraper.VCenterScraper
	logger     *slog.Logger
	conf       CollectorConfig
	collectors map[*helper.Matcher]prometheus.Collector
}

func NewVCCollector(conf *CollectorConfig, scraper *scraper.VCenterScraper, logger *slog.Logger) *VCCollector {
	if conf == nil {
		conf = DefaultCollectorConf()
	}

	collectors := map[*helper.Matcher]prometheus.Collector{}
	collectors[helper.NewMatcher("esx", "host")] = NewEsxCollector(scraper, *conf)
	collectors[helper.NewMatcher("ds", "datastore")] = NewDatastoreCollector(scraper, *conf)
	collectors[helper.NewMatcher("resourcepool", "rp")] = NewResourcePoolCollector(scraper, *conf)
	collectors[helper.NewMatcher("cluster", "host")] = NewClusterCollector(scraper, *conf)
	collectors[helper.NewMatcher("vm", "virtualmachine")] = NewVirtualMachineCollector(scraper, *conf)
	collectors[helper.NewMatcher("spod", "storagepod")] = NewStoragePodCollector(scraper, *conf)
	collectors[helper.NewMatcher("scraper")] = NewScraperCollector(scraper)

	return &VCCollector{
		scraper:    scraper,
		logger:     logger,
		conf:       *conf,
		collectors: collectors,
	}
}

func (c *VCCollector) GetMetricHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		filters := []string{}
		if f, ok := params["collect[]"]; ok {
			filters = append(filters, f...)
		}
		if f, ok := params["collect"]; ok {
			filters = append(filters, f...)
		}

		exclude := []string{}
		if f, ok := params["exclude[]"]; ok {
			exclude = append(exclude, f...)
		}
		if f, ok := params["exclude"]; ok {
			exclude = append(exclude, f...)
		}
		excludeMatcher := helper.NewMatcher(exclude...)
		c.logger.Debug(fmt.Sprintf("%s %s", r.Method, r.URL.Path), "filters", filters, "exclude", exclude)

		registry := prometheus.NewRegistry()
		if !c.conf.DisableExporterMetrics && !excludeMatcher.MatchAny("exporter_metrics", "exporter") {
			registry.MustRegister(
				collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
				collectors.NewGoCollector(),
			)
		}
		for matcher, collector := range c.collectors {
			if (len(filters) == 0 || slices.ContainsFunc(filters, matcher.Match)) && !excludeMatcher.MatchAny(matcher.Keywords...) {
				c.logger.Debug(fmt.Sprintf("register %s collector", matcher.First()))
				err := registry.Register(collector)
				if err != nil {
					c.logger.Error(fmt.Sprintf("Error registring %s collector for %s", matcher.First(), strings.Join(filters, ",")), "err", err.Error())
				}
			}
		}

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: c.conf.MaxRequests,
		})
		h.ServeHTTP(w, r)
	}
}
