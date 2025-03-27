package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/common/promslog"

	"github.com/prometheus/common/version"
	"github.com/sanderdescamps/govc_exporter/internal/collector"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
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
	ctx := context.Background()

	config := LoadConfig()
	if err := config.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid config: %v\n", err)
	}

	logger := promslog.New(config.PromlogConfig)
	logger.Info("Starting govc_exporter", "version", version.Version, "branch", version.Branch, "revision", version.GetRevision())
	logger.Info("Build context", "go", version.GoVersion, "platform", fmt.Sprintf("%s/%s", version.GoOS, version.GoArch), "user", version.BuildUser, "date", version.BuildDate, "tags", version.GetTags())

	if err := config.ScraperConfig.Validate(); err != nil {
		logger.Error("invalid scraper config", "err", err)
		os.Exit(1)
	}

	scraper, err := scraper.NewVCenterScraper(*config.ScraperConfig, logger)
	if err != nil {
		logger.Error("Failed to create VCenterScraper", "err", err)
		return
	}
	err = scraper.Start(ctx, logger)
	if err != nil {
		logger.Error("Failed to start VCenterScraper", "err", err)
		return
	}

	collector := collector.NewVCCollector(config.CollectorConfig, scraper, logger)

	server := &http.Server{
		Addr: config.ListenAddress,
	}

	http.HandleFunc(config.MetricPath, collector.GetMetricHandler(logger))
	if config.AllowManualRefresh {
		http.HandleFunc("/refresh/{sensor}", collector.GetRefreshHandler(logger))
	}
	if config.AllowDumps {
		http.HandleFunc("/dump", collector.GetDumpHandler(logger))
		http.HandleFunc("/dump/{sensor}", collector.GetDumpHandler(logger))
	}

	http.HandleFunc("/", defaultHandler(config.MetricPath))

	shutdownChan := make(chan bool, 1)

	go func() {
		logger.Info("Listening on", "address", config.ListenAddress)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server error", "err", err)
		}

		logger.Info("Stopped serving new connections.")
		scraper.Stop(ctx, logger)

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
