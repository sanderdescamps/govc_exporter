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

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/common/promslog"

	"github.com/intrinsec/govc_exporter/collector"
	"github.com/intrinsec/govc_exporter/collector/scraper"
	"github.com/prometheus/common/version"
)

func defaultHandler(metricsPath string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>govc Exporter</title></head>
			<body>
			<h1>govc Exporter</h1>
			<p><a href="` + metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	}
}

func main() {
	config := LoadConfig()
	logger := promslog.New(config.PromlogConfig)
	logger.Info("Starting govc_exporter", "version", version.Version, "branch", version.Branch, "revision", version.GetRevision())
	logger.Info("Build context", "go", version.GoVersion, "platform", fmt.Sprintf("%s/%s", version.GoOS, version.GoArch), "user", version.BuildUser, "date", version.BuildDate, "tags", version.GetTags())

	scraper, err := scraper.NewVCenterScraper(*config.ScraperConfig)
	if err != nil {
		logger.Error("Failed to create VCenterScraper", "err", err)
		return
	}
	err = scraper.Start(logger)
	if err != nil {
		logger.Error("Failed to start VCenterScraper", "err", err)
		return
	}

	collector := collector.NewVCCollector(*config.CollectorConfig, scraper, logger)

	server := &http.Server{
		Addr: config.ListenAddress,
	}

	http.HandleFunc(config.MetricPath, collector.GetMetricHandler())
	http.HandleFunc("/", defaultHandler(config.MetricPath))

	shutdownChan := make(chan bool, 1)

	go func() {
		logger.Info("Listening on", "address", config.ListenAddress)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server error", "err", err)
		}

		logger.Info("Stopped serving new connections.")
		scraper.Stop(logger)

		shutdownChan <- true
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP shutdown error", "err", err)
	}

	<-shutdownChan
	logger.Info("Graceful shutdown complete.")
}

func string2bool(s string) bool {
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	return false
}
