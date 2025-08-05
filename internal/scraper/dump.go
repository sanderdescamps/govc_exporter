package scraper

import (
	"encoding/json"
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

		if len(refTypes) < 1 {
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

		for _, refType := range refTypes {
			jsonByte, err := scraper.DB.JsonDump(ctx, refType)
			if err != nil {
				logger.Error("Failed to create dump", "err", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"msg":    "Failed to create dump",
					"err":    err.Error(),
					"status": http.StatusInternalServerError,
				})
				return
			}
			filePath := path.Join(dirPath, fmt.Sprintf("%s.json", refType.String()))
			os.WriteFile(filePath, jsonByte, os.ModePerm)
		}

		logger.Info(fmt.Sprintf("Dump successful. Check %s for results", dirPath))
	})
}
