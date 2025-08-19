package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
)

func GetDumpHandler(scraper VCenterScraper, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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

		refTypes := []objects.ManagedObjectTypes{}
		perfMetricTypes := []objects.PerfMetricTypes{}

		if helper.NewMatcher("cluster").MatchAny(include...) {
			refTypes = append(refTypes, objects.ManagedObjectTypesCluster)
		}
		if helper.NewMatcher("compute_resource", "computeresource", "compresource").MatchAny(include...) {
			refTypes = append(refTypes, objects.ManagedObjectTypesComputeResource)
		}
		if helper.NewMatcher("dc", "datacenter").MatchAny(include...) {
			refTypes = append(refTypes, objects.ManagedObjectTypesDatacenter)
		}
		if helper.NewMatcher("ds", "datastore").MatchAny(include...) {
			refTypes = append(refTypes, objects.ManagedObjectTypesDatastore)
		}
		if helper.NewMatcher("folder").MatchAny(include...) {
			refTypes = append(refTypes, objects.ManagedObjectTypesFolder)
		}
		if helper.NewMatcher("host", "esx").MatchAny(include...) {
			refTypes = append(refTypes, objects.ManagedObjectTypesHost)
		}
		if helper.NewMatcher("resource_pool", "rpool", "respool").MatchAny(include...) {
			refTypes = append(refTypes, objects.ManagedObjectTypesResourcePool)
		}
		if helper.NewMatcher("storagepod", "storage_pod", "datastore_cluster", "datastorecluster").MatchAny(include...) {
			refTypes = append(refTypes, objects.ManagedObjectTypesStoragePod)
		}
		if helper.NewMatcher("tags", "tag").MatchAny(include...) {
			refTypes = append(refTypes, objects.ManagedObjectTypesTagSet)
		}
		if helper.NewMatcher("vm", "virtualmachine", "virtual_machine").MatchAny(include...) {
			refTypes = append(refTypes, objects.ManagedObjectTypesVirtualMachine)
		}

		if helper.NewMatcher("perfvm", "perf_virtual_machine").MatchAny(include...) {
			perfMetricTypes = append(perfMetricTypes, objects.PerfMetricTypesVirtualMachine)
		}
		if helper.NewMatcher("perfesx", "perf_esx", "perfhost", "perf_host").MatchAny(include...) {
			perfMetricTypes = append(perfMetricTypes, objects.PerfMetricTypesHost)
		}

		if len(refTypes) < 1 && len(perfMetricTypes) < 1 {
			logger.Warn("Failed to create dump. Invalid sensor type", "type", include)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"msg":    fmt.Sprintf("Failed to create dump. Invalid sensor type %v", include),
				"status": http.StatusBadRequest,
			})
			return
		}
		logger.Info("Creating dump of DB tables.", "ref_types", func() string {
			result := []string{}
			for _, t := range refTypes {
				result = append(result, t.String())
			}
			return strings.Join(result, ";")
		}())

		dirPath, err := makeDumpDir()
		if err != nil {
			logger.Warn("Failed to create dump dir", "err", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"msg":    "Failed to create dump",
				"status": http.StatusInternalServerError,
			})
			return
		}

		for _, refType := range refTypes {
			jsonByte, err := scraper.DB.JsonDump(ctx, refType)
			if err != nil {
				logger.Error("Failed to create dump", "err", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"msg":    "Failed to create dump",
					"status": http.StatusInternalServerError,
				})
				return
			}
			filePath := path.Join(dirPath, fmt.Sprintf("%s.json", refType.String()))
			os.WriteFile(filePath, jsonByte, os.ModePerm)
		}

		pmDumps, err := scraper.MetricsDB.JsonDump(ctx, perfMetricTypes...)
		if err != nil {
			logger.Error("Failed to create perfdump", "err", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"msg":    "Failed to create dump",
				"status": http.StatusInternalServerError,
			})
			return
		}
		for ref, pmDump := range pmDumps {
			filePath := path.Join(dirPath, fmt.Sprintf("perf-%s.json", ref.Value))
			os.WriteFile(filePath, pmDump, os.ModePerm)
		}

		logger.Info(fmt.Sprintf("Dump created successful. Check %s for results.", dirPath))
		logger.Warn(fmt.Sprintf("Cleanup %s when not needed anymore", dirPath))
		json.NewEncoder(w).Encode(map[string]any{
			"msg":    "Dump has been created on server. Make sure to cleanup dir when not needed anymore",
			"dir":    dirPath,
			"status": http.StatusOK,
		})
	})
}

func makeDumpDir() (string, error) {
	for i := 0; true; i++ {
		dirPath := fmt.Sprintf("./dumps/%s_%d", time.Now().Format("20060102"), i)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			err := os.MkdirAll(dirPath, 0775)
			if errors.Is(err, os.ErrPermission) {
				//continue bellow, create a dir in temp dir
				break
			} else if err != nil {
				return "", fmt.Errorf("failed to create %s dir: %v", dirPath, err)
			}
			return dirPath, nil
		}
	}

	dirPath, err := os.MkdirTemp("", fmt.Sprintf("govc-dump-%s-", time.Now().Format("20060102")))
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %v", err)
	}
	return dirPath, nil
}
