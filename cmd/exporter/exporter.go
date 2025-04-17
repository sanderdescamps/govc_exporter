package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/prometheus/common/promslog"

	"github.com/prometheus/common/version"
	"github.com/sanderdescamps/govc_exporter/internal/collector"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

func defaultHandler(metricsPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>govc Exporter</title></head>
			<body>
			<h1>govc Exporter</h1>
			<p><a href="` + metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
}

func main() {
	run()
}

func run() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	config := LoadConfig()
	if err := config.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid config: %v\n", err)
		os.Exit(1)
	}

	logger := promslog.New(config.PromlogConfig)
	logger.Info("Starting govc_exporter", "version", version.Version, "branch", version.Branch, "revision", version.GetRevision())
	logger.Info("Build context", "go", version.GoVersion, "platform", fmt.Sprintf("%s/%s", version.GoOS, version.GoArch), "date", version.BuildDate, "tags", version.GetTags())

	if os.Getenv("GOMEMLIMIT") == "" && config.MemoryLimitMB > 0 {
		logger.Debug(fmt.Sprintf("Set memory limit to %dMiB", config.MemoryLimitMB))
		debug.SetMemoryLimit(config.MemoryLimitMB * 1 << 20)
	}
	logger.Debug(fmt.Sprintf("Memory limit set to %dMiB", debug.SetMemoryLimit(-1)*1>>20))

	//Scraper
	ctxScraper := context.WithValue(ctx, scraper.ContextKeyScraperLogger{}, logger)
	scrap, err := scraper.NewVCenterScraper(ctxScraper, *config.ScraperConfig)
	if err != nil {
		logger.Error("Failed to create VCenterScraper", "err", err)
		return
	}
	err = scrap.Start(ctxScraper)
	if err != nil {
		logger.Error("Failed to start VCenterScraper", "err", err)
		return
	}

	//Collector
	ctxCollector := context.WithValue(ctx, collector.ContextKeyCollectorLogger{}, logger)
	coll := collector.NewVCCollector(ctxCollector, config.CollectorConfig, scrap)

	//Server
	server := &http.Server{
		Addr: config.ListenAddress,
		BaseContext: func(l net.Listener) context.Context {
			return ctxCollector
		},
	}

	http.Handle(config.MetricPath, coll.GetMetricHandler())
	// http.Handle("/pprof/metrics", coll.GetMetricHandlerCPUPProf("metrics"))
	if config.AllowManualRefresh {
		http.Handle("/refresh/{sensor}", coll.GetRefreshHandler())
	}
	if config.AllowDumps {
		http.Handle("/dump", coll.GetDumpHandler())
		http.Handle("/dump/{sensor}", coll.GetDumpHandler())
	}
	http.Handle("/", defaultHandler(config.MetricPath))

	// make it a goroutine
	go server.ListenAndServe()
	logger.Info("Listening on", "address", config.ListenAddress)

	// listen for the interrupt signal
	<-ctx.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	shutdown := make(chan error)

	// stop the server
	go func() {
		err := server.Shutdown(stopCtx)
		if err != nil {
			shutdown <- err
		}
		scrap.Stop(ctx)
		shutdown <- nil
	}()
	select {
	case err := <-shutdown:
		if err != nil {
			log.Fatalf("could not shutdown: %v\n", err)
			os.Exit(1)
		}
	case err := <-stopCtx.Done():
		log.Fatalf("shutdown timeout: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Shutdown complete.")
}
