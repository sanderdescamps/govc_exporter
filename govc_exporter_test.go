// Copyright 2017 The Prometheus Authors
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
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/intrinsec/govc_exporter/collector"
	"github.com/intrinsec/govc_exporter/collector/scraper"
	"github.com/prometheus/common/version"
	"github.com/prometheus/procfs"
)

var (
	binary = filepath.Join(os.Getenv("GOPATH"), "bin/govc_exporter")
)

const (
	address = "localhost:19123"
)

func TestGeneralExporter(t *testing.T) {
	collectorConf := collector.CollectorConfig{
		UseIsecSpecifics:       false,
		CollectVMNetworks:      true,
		CollectVMDisks:         true,
		DisableExporterMetrics: false,
		MaxRequests:            4,
	}
	scraperConf := scraper.ScraperConfig{
		Endpoint:                         "https://localhost:8989",
		Username:                         "testuser",
		Password:                         "testpass",
		HostCollectorEnabled:             true,
		HostMaxAgeSec:                    60,
		HostRefreshIntervalSec:           20,
		ClusterCollectorEnabled:          true,
		ClusterMaxAgeSec:                 60,
		ClusterRefreshIntervalSec:        20,
		VirtualMachineCollectorEnabled:   true,
		VirtualMachineMaxAgeSec:          60,
		VirtualMachineRefreshIntervalSec: 20,
		DatastoreCollectorEnabled:        true,
		DatastoreMaxAgeSec:               60,
		DatastoreRefreshIntervalSec:      20,
		SpodCollectorEnabled:             true,
		SpodMaxAgeSec:                    60,
		SpodRefreshIntervalSec:           20,
		ResourcePoolCollectorEnabled:     true,
		ResourcePoolMaxAgeSec:            60,
		ResourcePoolRefreshIntervalSec:   20,
		OnDemandCacheMaxAge:              60,
		CleanIntervalSec:                 5,
		ClientPoolSize:                   5,
	}

	metricsPath := "/metrics"
	listenAddress := ":9752"
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// ctx, cancel := context.WithCancelCause(context.Background())
	// collector := collector.NewVCCollector(collectorConf, logger)
	// collector.Start(ctx)

	// http.HandleFunc(metricsPath, collector.GetMetricHandler())
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte(`<html>
	// 		<head><title>govc Exporter</title></head>
	// 		<body>
	// 		<h1>govc Exporter</h1>
	// 		<p><a href="` + metricsPath + `">Metrics</a></p>
	// 		</body>
	// 		</html>`))
	// })

	// logger.Info("Listening on", "address", listenAddress)
	// server := &http.Server{
	// 	Addr: listenAddress,
	// }

	scraper, err := scraper.NewVCenterScraper(scraperConf)
	if err != nil {
		logger.Error("Failed to create VCenterScraper", "err", err.Error())
	}
	scraper.Start(logger)

	collector := collector.NewVCCollector(collectorConf, scraper, logger)

	logger.Info("Starting govc_exporter", "version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())

	server := &http.Server{
		Addr: listenAddress,
	}

	http.HandleFunc(metricsPath, collector.GetMetricHandler())
	http.HandleFunc("/", defaultHandler(metricsPath))

	err = server.ListenAndServe()
	if err != nil {
		logger.Error("Exporter failed", "err", err)
		os.Exit(1)
	}
}

func TestFileDescriptorLeak(t *testing.T) {
	if _, err := os.Stat(binary); err != nil {
		t.Skipf("govc_exporter binary not available, try to run `make build` first: %s", err)
	}
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		t.Skipf("proc filesystem is not available, but currently required to read number of open file descriptors: %s", err)
	}
	if _, err := fs.Stat(); err != nil {
		t.Errorf("unable to read process stats: %s", err)
	}

	exporter := exec.Command(binary, "--web.listen-address", address)
	test := func(pid int) error {
		if err := queryExporter(address); err != nil {
			return err
		}
		proc, err := procfs.NewProc(pid)
		if err != nil {
			return err
		}
		fdsBefore, err := proc.FileDescriptors()
		if err != nil {
			return err
		}
		for i := 0; i < 5; i++ {
			if err := queryExporter(address); err != nil {
				return err
			}
		}
		fdsAfter, err := proc.FileDescriptors()
		if err != nil {
			return err
		}
		if want, have := len(fdsBefore), len(fdsAfter); want != have {
			return fmt.Errorf("want %d open file descriptors after metrics scrape, have %d", want, have)
		}
		return nil
	}

	if err := runCommandAndTests(exporter, address, test); err != nil {
		t.Error(err)
	}
}

func TestHandlingOfDuplicatedMetrics(t *testing.T) {
	if _, err := os.Stat(binary); err != nil {
		t.Skipf("govc_exporter binary not available, try to run `make build` first: %s", err)
	}

	exporter := exec.Command(binary, "--web.listen-address", address)
	test := func(_ int) error {
		return queryExporter(address)
	}

	if err := runCommandAndTests(exporter, address, test); err != nil {
		t.Error(err)
	}
}

func queryExporter(address string) error {
	resp, err := http.Get(fmt.Sprintf("http://%s/metrics", address))
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}
	if want, have := http.StatusOK, resp.StatusCode; want != have {
		return fmt.Errorf("want /metrics status code %d, have %d. Body:\n%s", want, have, b)
	}
	return nil
}

func runCommandAndTests(cmd *exec.Cmd, address string, fn func(pid int) error) error {
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %s", err)
	}
	time.Sleep(50 * time.Millisecond)
	for i := 0; i < 10; i++ {
		if err := queryExporter(address); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
		if cmd.Process == nil || i == 9 {
			return fmt.Errorf("can't start command")
		}
	}

	errc := make(chan error)
	go func(pid int) {
		errc <- fn(pid)
	}(cmd.Process.Pid)

	err := <-errc
	if cmd.Process != nil {
		cmd.Process.Kill()
	}
	return err
}
