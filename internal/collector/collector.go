package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

// Namespace defines the common namespace to be used by all metrics.
const namespace = "govc"

type VCCollector struct {
	scraper *scraper.VCenterScraper
	// logger     *slog.Logger
	conf       Config
	collectors map[*helper.Matcher]prometheus.Collector
}

func NewVCCollector(conf *Config, scraper *scraper.VCenterScraper, logger *slog.Logger) *VCCollector {
	if conf == nil {
		conf = DefaultCollectorConf()
	}

	collectors := map[*helper.Matcher]prometheus.Collector{}
	collectors[helper.NewMatcher("esx", "host")] = NewEsxCollector(scraper, *conf)
	collectors[helper.NewMatcher("perfhost", "esx", "host")] = NewEsxPerfCollector(scraper, *conf)
	collectors[helper.NewMatcher("ds", "datastore")] = NewDatastoreCollector(scraper, *conf)
	collectors[helper.NewMatcher("resourcepool", "rp")] = NewResourcePoolCollector(scraper, *conf)
	collectors[helper.NewMatcher("cluster", "host")] = NewClusterCollector(scraper, *conf)
	collectors[helper.NewMatcher("vm", "virtualmachine")] = NewVirtualMachineCollector(scraper, *conf)
	collectors[helper.NewMatcher("perfvm", "vm", "virtualmachine")] = NewVMPerfCollector(scraper, *conf)
	collectors[helper.NewMatcher("spod", "storagepod")] = NewStoragePodCollector(scraper, *conf)
	collectors[helper.NewMatcher("scraper")] = NewScraperCollector(scraper)

	return &VCCollector{
		scraper: scraper,
		// logger:     logger,
		conf:       *conf,
		collectors: collectors,
	}
}

func (c *VCCollector) GetMetricHandler(logger *slog.Logger) func(w http.ResponseWriter, r *http.Request) {
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
		logger.Debug(fmt.Sprintf("%s %s", r.Method, r.URL.Path), "filters", filters, "exclude", exclude)

		registry := prometheus.NewRegistry()
		if !c.conf.DisableExporterMetrics && !excludeMatcher.MatchAny("exporter_metrics", "exporter") {
			registry.MustRegister(
				collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
				collectors.NewGoCollector(),
			)
		}
		for matcher, collector := range c.collectors {
			if (len(filters) == 0 || slices.ContainsFunc(filters, matcher.Match)) && !excludeMatcher.MatchAny(matcher.Keywords...) {
				logger.Debug(fmt.Sprintf("register %s collector", matcher.First()))
				err := registry.Register(collector)
				if err != nil {
					logger.Error(fmt.Sprintf("Error registring %s collector for %s", matcher.First(), strings.Join(filters, ",")), "err", err.Error())
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

func (c *VCCollector) GetRefreshHandler(logger *slog.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sensorReq := r.PathValue("sensor")

		sensor := c.scraper.GetSensorRefreshByName(sensorReq)
		if sensor != nil {
			sensorName := sensor.Name()
			logger.Info("Trigger manual refresh", "sensor_name", sensorName)
			err := sensor.Refresh(context.Background(), logger)
			if err == nil {
				logger.Info("Refresh successfull", "sensor_name", sensorName)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]any{
					"msg":         "refresh successfull",
					"sensor_name": sensorName,
					"status":      200,
				})
			} else {
				logger.Warn("Failed to refresh sensor", "err", err.Error(), "sensor_name", sensorName)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"msg":         "Failed to refresh sensor",
					"sensor_name": sensorName,
					"err":         err.Error(),
					"status":      http.StatusInternalServerError,
				})
			}
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"msg":    fmt.Sprintf("Invalid sensor %s", sensorReq),
				"status": http.StatusBadRequest,
			})
		}
	}
}

func (c *VCCollector) GetDumpHandler(logger *slog.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		include := []string{}
		sensorReq := r.PathValue("sensor")
		if sensorReq != "" {
			include = append(include, sensorReq)
		} else {
			params := r.URL.Query()
			if f, ok := params["collect[]"]; ok {
				include = append(include, f...)
			} else if f, ok := params["collect"]; ok {
				include = append(include, f...)
			} else {
				include = append(include, "all")
			}

		}

		sensors := []scraper.Sensor{}
		for _, sensor := range c.scraper.SensorList() {
			match := false
			for _, i := range include {
				if sensor.Match(i) {
					match = true
					break
				}
			}
			if match {
				sensors = append(sensors, sensor)
			}
		}

		if len(sensors) < 1 {
			logger.Warn("Failed to create dump. Invalid sensor type", "type", include)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"msg":    fmt.Sprintf("Failed to create dump. Invalid sensor type %v", include),
				"status": http.StatusBadRequest,
			})
			return
		}
		logger.Info("Creating dump of sensors objects.", "sensors", func() []string {
			result := []string{}
			for _, s := range sensors {
				result = append(result, s.Name())
			}
			return result
		})

		dirPath := ""
		for i := 0; true; i++ {
			dirPath = fmt.Sprintf("./dumps/%s_%d", time.Now().Format("20060201"), i)
			if _, err := os.Stat(dirPath); os.IsNotExist(err) {
				logger.Debug("Create dump path", "path", dirPath)
				err := os.MkdirAll(dirPath, 0775)
				if err != nil {
					logger.Warn("Failed to create dump", "err", err)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]any{
						"msg":    "Failed to create dump",
						"err":    err.Error(),
						"status": http.StatusInternalServerError,
					})
					return
				}
				break
			}
		}

		for _, sensor := range sensors {

			jsonMap, err := sensor.GetAllJsons()
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"msg":    "Failed to create dump",
					"err":    err.Error(),
					"status": http.StatusInternalServerError,
				})
				return
			}
			for name, jsonString := range jsonMap {
				filePath := path.Join(dirPath, fmt.Sprintf("%s-%s.json", sensor.Name(), name))
				os.WriteFile(filePath, jsonString, os.ModePerm)
			}
		}
		logger.Info(fmt.Sprintf("Dump successful. Check %s for results", dirPath))
	}
}

// func getSensorMatchMap(scraper *scraper.VCenterScraper) map[*helper.Matcher]interface {
// 	GetKind() string
// 	GetAllJsons() (map[string][]byte, error)
// } {
// 	return map[*helper.Matcher]interface {
// 		GetKind() string
// 		GetAllJsons() (map[string][]byte, error)
// 	}{
// 		helper.NewMatcher("cl", "cluster"):                                                     scraper.Cluster,
// 		helper.NewMatcher("cr", "compute_resource", "computeresource"):                         scraper.ComputeResources,
// 		helper.NewMatcher("ds", "datastore", "datastores"):                                     scraper.Datastore,
// 		helper.NewMatcher("h", "esx", "host"):                                                  scraper.Host,
// 		helper.NewMatcher("rp", "resource_pool", "respool", "repool"):                          scraper.ResourcePool,
// 		helper.NewMatcher("sp", "storage_pod", "storagepod", "dscluster", "datastore_cluster"): scraper.SPOD,
// 		helper.NewMatcher("t", "tag", "tags"):                                                  scraper.Tags,
// 		helper.NewMatcher("vm", "vms", "virtualmachine"):                                       scraper.VM,
// 	}
// }
