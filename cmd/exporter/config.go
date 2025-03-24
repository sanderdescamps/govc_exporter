package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/sanderdescamps/govc_exporter/internal/collector"
	"github.com/sanderdescamps/govc_exporter/internal/scraper"
)

type Config struct {
	ListenAddress      string
	MetricPath         string
	AllowDumps         bool
	AllowManualRefresh bool
	ScraperConfig      *scraper.Config
	CollectorConfig    *collector.Config
	PromlogConfig      *promslog.Config
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

func LoadConfig() Config {
	cfg := Config{
		ScraperConfig:   &scraper.Config{},
		CollectorConfig: &collector.Config{},
		PromlogConfig:   &promslog.Config{},
	}

	a := kingpin.New(filepath.Base(os.Args[0]), "Prometheus vCenter exporter")

	flag.AddFlags(a, cfg.PromlogConfig)
	a.Version(version.Print("govc_exporter"))
	a.HelpFlag.Short('h')

	//web
	a.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9752").StringVar(&cfg.ListenAddress)
	a.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").StringVar(&cfg.MetricPath)
	a.Flag("web.max-requests", "Maximum number of parallel scrape requests. Use 0 to disable.").Default("40").IntVar(&cfg.CollectorConfig.MaxRequests)
	a.Flag("web.manual-refresh", "Enable /refresh/{sensor} path to trigger a refresh of a sensor.").Default("false").BoolVar(&cfg.AllowManualRefresh)
	a.Flag("web.allow-dumps", "Enable /dump path to trigger a dump of the cache data in ./dumps folder on server side. Only enable for debugging.").Default("false").BoolVar(&cfg.AllowDumps)

	//collector
	a.Flag("web.disable-exporter-metrics", "Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).").BoolVar(&cfg.CollectorConfig.DisableExporterMetrics)
	a.Flag("collector.intrinsec", "Enable intrinsec specific features").Default("false").BoolVar(&cfg.CollectorConfig.UseIsecSpecifics)

	//collector.cluster
	a.Flag("collector.cluster.tag_label", "List of vmware tag categories to collect which will be added as label in metrics").StringsVar(&cfg.CollectorConfig.ClusterTagLabels)

	//collector.datastore
	a.Flag("collector.datastore.tag_label", "List of vmware tag categories to collect which will be added as label in metrics").StringsVar(&cfg.CollectorConfig.DatastoreTagLabels)

	//collector.host
	a.Flag("collector.host.storage", "Collect host storage metrics").Default("false").BoolVar(&cfg.CollectorConfig.HostStorageMetrics)
	a.Flag("collector.host.tag_label", "List of vmware tag categories which will be added as label in metrics").StringsVar(&cfg.CollectorConfig.HostTagLabels)

	//collector.repool
	a.Flag("collector.repool.tag_label", "List of tag categories which will be added as label in metrics").StringsVar(&cfg.CollectorConfig.ResourcePoolTagLabels)

	//collector.spod
	a.Flag("collector.spod.tag_label", "List of vmware tag categories to collect which will be added as label in metrics").StringsVar(&cfg.CollectorConfig.StoragePodTagLabels)

	//collector.vm
	a.Flag("collector.vm.tag_label", "List of vmware tag categories to collect which will be added as label in metrics").StringsVar(&cfg.CollectorConfig.VMTagLabels)

	//scraper
	a.Flag("scraper.vc.url", "vc api username").Envar("VC_URL").Required().StringVar(&cfg.ScraperConfig.Endpoint)
	a.Flag("scraper.vc.username", "vc api username").Envar("VC_USERNAME").Required().StringVar(&cfg.ScraperConfig.Username)
	a.Flag("scraper.vc.password", "vc api password").Envar("VC_PASSWORD").Required().StringVar(&cfg.ScraperConfig.Password)
	a.Flag("scraper.client_pool_size", "number of simultanious requests to vCenter api").Default("5").IntVar(&cfg.ScraperConfig.ClientPoolSize)
	a.Flag("scraper.clean_interval", "interval the scraper cleans up old data").Default("5s").DurationVar(&cfg.ScraperConfig.CleanInterval)

	//scraper.cluster
	a.Flag("scraper.cluster", "Enable cluster scraper").Default("True").BoolVar(&cfg.ScraperConfig.ClusterScraperEnabled)
	a.Flag("scraper.cluster.max_age", "time in seconds clusters are cached").Default("5m").DurationVar(&cfg.ScraperConfig.ClusterMaxAge)
	a.Flag("scraper.cluster.refresh_interval", "interval clusters are refreshed").Default("25s").DurationVar(&cfg.ScraperConfig.ClusterRefreshInterval)

	//scraper.compute_resource
	a.Flag("scraper.compute_resource", "Enable compute_resource scraper").Default("True").BoolVar(&cfg.ScraperConfig.ComputeResourceScraperEnabled)
	a.Flag("scraper.compute_resource.max_age", "time in seconds clusters are cached").Default("5m").DurationVar(&cfg.ScraperConfig.ComputeResourceMaxAge)
	a.Flag("scraper.compute_resource.refresh_interval", "interval clusters are refreshed").Default("25s").DurationVar(&cfg.ScraperConfig.ComputeResourceRefreshInterval)

	//scraper.datastore
	a.Flag("scraper.ds", "Enable datastore scraper").Default("True").BoolVar(&cfg.ScraperConfig.DatastoreScraperEnabled)
	a.Flag("scraper.datastore.max_age", "time in seconds datastores are cached").Default("2m").DurationVar(&cfg.ScraperConfig.DatastoreMaxAge)
	a.Flag("scraper.datastore.refresh_interval", "interval datastores are refreshed").Default("55s").DurationVar(&cfg.ScraperConfig.DatastoreRefreshInterval)

	//scraper.host
	a.Flag("scraper.host", "Enable host scraper").Default("True").BoolVar(&cfg.ScraperConfig.HostScraperEnabled)
	a.Flag("scraper.host.max_age", "time in seconds hosts are cached").Default("1m").DurationVar(&cfg.ScraperConfig.HostMaxAge)
	a.Flag("scraper.host.refresh_interval", "interval hosts are refreshed").Default("25s").DurationVar(&cfg.ScraperConfig.HostRefreshInterval)

	//scraper.repool
	a.Flag("scraper.repool", "Enable resource pool scraper").Default("True").BoolVar(&cfg.ScraperConfig.ResourcePoolScraperEnabled)
	a.Flag("scraper.repool.max_age", "time in seconds resource pools are cached").Default("2m").DurationVar(&cfg.ScraperConfig.ResourcePoolMaxAge)
	a.Flag("scraper.repool.refresh_interval", "interval resource pools are refreshed").Default("55s").DurationVar(&cfg.ScraperConfig.ResourcePoolRefreshInterval)

	//scraper.spod
	a.Flag("scraper.spod", "Enable datastore cluster scraper").Default("True").BoolVar(&cfg.ScraperConfig.SpodScraperEnabled)
	a.Flag("scraper.spod.max_age", "time in seconds spods are cached").Default("2m").DurationVar(&cfg.ScraperConfig.SpodMaxAge)
	a.Flag("scraper.spod.refresh_interval", "interval spods are refreshed").Default("55s").DurationVar(&cfg.ScraperConfig.SpodRefreshInterval)

	//scraper.tags
	a.Flag("scraper.tags", "Collect tags").Default("True").BoolVar(&cfg.ScraperConfig.TagsScraperEnabled)
	a.Flag("scraper.tags.max_age", "time in seconds tags are cached").Default("10m").DurationVar(&cfg.ScraperConfig.TagsMaxAge)
	a.Flag("scraper.tags.refresh_interval", "interval tags are refreshed").Default("290s").DurationVar(&cfg.ScraperConfig.TagsRefreshInterval)

	//scraper.vm
	a.Flag("scraper.vm", "Enable virtualmachine scraper").Default("True").BoolVar(&cfg.ScraperConfig.VirtualMachineScraperEnabled)
	a.Flag("scraper.vm.max_age", "time in seconds vm's are cached").Default("2m").DurationVar(&cfg.ScraperConfig.VirtualMachineMaxAge)
	a.Flag("scraper.vm.refresh_interval", "interval vm's are refreshed").Default("55s").DurationVar(&cfg.ScraperConfig.VirtualMachineRefreshInterval)
	a.Flag("collector.vm.disk", "Collect extra vm disk metrics").Default("false").BoolVar(&cfg.CollectorConfig.VMAdvancedStorageMetrics)
	a.Flag("collector.vm.network", "Collect extra vm network metrics").Default("false").BoolVar(&cfg.CollectorConfig.VMAdvancedNetworkMetrics)

	//scraper.on_demand
	a.Flag("scraper.on_demand.max_age", "Time in seconds the scraper keeps all non-cache data. Used when no other sensor is available").Default("5m").DurationVar(&cfg.ScraperConfig.OnDemandCacheMaxAge)

	if _, err := a.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing commandline arguments: %v\n", err)
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	cfg.ScraperConfig.TagsCategoryToCollect = dedup(mergeLists(
		cfg.CollectorConfig.ClusterTagLabels,
		cfg.CollectorConfig.DatastoreTagLabels,
		cfg.CollectorConfig.HostTagLabels,
		cfg.CollectorConfig.ResourcePoolTagLabels,
		cfg.CollectorConfig.StoragePodTagLabels,
		cfg.CollectorConfig.VMTagLabels,
	))

	return cfg
}
