package config

import (
	"fmt"
	"strings"

	"github.com/prometheus/common/promslog"
)

type Config struct {
	ListenAddress      string
	MetricPath         string
	AllowDumps         bool
	AllowManualRefresh bool
	ScraperConfig      ScraperConfig
	CollectorConfig    CollectorConfig
	PromlogConfig      promslog.Config

	// MemoryLimit is the memory limit in MB for the process.
	// It is set to 0 if not specified.
	MemoryLimitMB int64
}

func (c Config) Validate() error {
	if !strings.HasPrefix(c.MetricPath, "/") {
		return fmt.Errorf("MetricPath must start with a '/'")
	}

	var err error
	if err = c.CollectorConfig.Validate(); err != nil {
		return fmt.Errorf("collector: %s", err.Error())
	}

	if err = c.ScraperConfig.Validate(); err != nil {
		return fmt.Errorf("scraper: %s", err.Error())
	}
	return nil
}

func DefaultConfig() Config {
	return Config{
		ScraperConfig:      DefaultScraperConfig(),
		PromlogConfig:      promslog.Config{},
		CollectorConfig:    DefaultCollectorConf(),
		ListenAddress:      ":9752",
		AllowDumps:         false,
		AllowManualRefresh: false,
		MetricPath:         "/metrics",
		MemoryLimitMB:      0,
	}
}
