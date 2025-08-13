package collector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

// Namespace defines the common namespace to be used by all metrics.
const namespace = "govc"

type VCCollector struct {
	scraper *scraper.VCenterScraper
	// logger     *slog.Logger
	conf       config.Config
	collectors map[*helper.Matcher]prometheus.Collector
}

func NewVCCollector(ctx context.Context, conf config.Config, scraper *scraper.VCenterScraper) *VCCollector {

	collectors := map[*helper.Matcher]prometheus.Collector{}
	collectors[helper.NewMatcher("esx", "host")] = NewEsxCollector(scraper, conf.CollectorConfig)
	collectors[helper.NewMatcher("ds", "datastore")] = NewDatastoreCollector(scraper, conf.CollectorConfig)
	collectors[helper.NewMatcher("resourcepool", "rp", "rpool")] = NewResourcePoolCollector(scraper, conf.CollectorConfig)
	collectors[helper.NewMatcher("cluster", "clu")] = NewClusterCollector(scraper, conf.CollectorConfig)
	collectors[helper.NewMatcher("vm", "virtualmachine")] = NewVirtualMachineCollector(scraper, conf.CollectorConfig)

	if conf.ScraperConfig.VirtualMachinePerf.Enabled {
		collectors[helper.NewMatcher("perfvm", "perf-vm")] = NewVMPerfCollector(scraper, conf.CollectorConfig)
	}

	if conf.ScraperConfig.HostPerf.Enabled {
		collectors[helper.NewMatcher("perfhost", "perfesx", "perf-host", "perf-esx")] = NewEsxPerfCollector(scraper, conf.CollectorConfig)
	}

	collectors[helper.NewMatcher("spod", "storagepod")] = NewStoragePodCollector(scraper, conf.CollectorConfig)
	collectors[helper.NewMatcher("scraper")] = NewScraperCollector(scraper)

	return &VCCollector{
		scraper:    scraper,
		conf:       conf,
		collectors: collectors,
	}
}

func (c *VCCollector) GetMetricHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		logger.Debug(fmt.Sprintf("%s %s", r.Method, r.URL.Path), "filters", filters, "exclude", exclude)

		registry := prometheus.NewRegistry()
		if !c.conf.CollectorConfig.DisableExporterMetrics && !excludeMatcher.MatchAny("exporter_metrics", "exporter") {
			registry.MustRegister(
				collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
				collectors.NewGoCollector(),
			)
		}
		found := false
		for matcher, collector := range c.collectors {
			if (len(filters) == 0 || slices.ContainsFunc(filters, matcher.Match)) && !excludeMatcher.MatchAny(matcher.Keywords...) {
				logger.Debug(fmt.Sprintf("register %s collector", matcher.First()))

				err := registry.Register(collector)
				if err != nil {
					logger.Error(fmt.Sprintf("Error registring %s collector for %s", matcher.First(), strings.Join(filters, ",")), "err", err.Error())

				}
				found = true
			}
		}
		if !found {
			logger.Warn("No sensor found for filter", "filter", filters)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{
				"msg":    fmt.Sprintf("No sensor found for filter for %v", filters),
				"err":    "No sensor found for filter",
				"status": http.StatusNotFound,
			})
		}

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: c.conf.CollectorConfig.MaxRequests,
			ErrorLog:            slog.NewLogLogger(logger.Handler(), slog.LevelError),
		})
		h.ServeHTTP(w, r)
	})
}

func (c *VCCollector) GetRefreshHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sensorName := r.PathValue("sensor")

		logger.Info("trigger manual refresh", "sensor_name", sensorName)
		err := c.scraper.TriggerSensorRefreshByName(ctx, sensorName)

		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, scraper.ErrSensorNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{
				"msg":         "sensor not found",
				"sensor_name": sensorName,
				"status":      404,
			})
		} else if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]any{
				"msg":         "Unknown error",
				"sensor_name": sensorName,
				"status":      502,
			})
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"msg":         "Refresh on sensor succesfully triggered",
				"sensor_name": sensorName,
				"status":      200,
			})
		}
	})
}
