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
	"time"

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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// if cpuPProf := os.Getenv("CPU_PPROF"); cpuPProf != "" {
	// 	if pprofEnabled, err := strconv.ParseBool(cpuPProf); err == nil && pprofEnabled {

	// 		pprofFile, err := createPProfFile("global-CPU")
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		defer pprofFile.Close()

	// 		if err := pprof.StartCPUProfile(pprofFile); err != nil {
	// 			panic(err)
	// 		}
	// 		defer pprof.StopCPUProfile()
	// 	}

	// }
	run(ctx)
}

func run(ctx context.Context) {

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
	scrap, err := scraper.NewVCenterScraper(ctxScraper, *config.ScraperConfig, logger)
	if err != nil {
		logger.Error("Failed to create VCenterScraper", "err", err)
		return
	}
	err = scrap.Start(ctxScraper, logger)
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
	if config.AllowManualRefresh {
		http.Handle("/refresh/{sensor}", coll.GetRefreshHandler())
	}
	if config.AllowDumps {
		http.Handle("/dump", scraper.GetDumpHandler(*scrap, logger))
		http.Handle("/dump/{sensor}", scraper.GetDumpHandler(*scrap, logger))
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
		scrap.Stop(ctx, logger)
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
