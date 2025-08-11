package main

import (
	"fmt"
	"os"
	"path/filepath"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/sanderdescamps/govc_exporter/internal/config"
	"github.com/sanderdescamps/govc_exporter/internal/helper"
)

func LoadConfig() config.Config {
	cfg := config.DefaultConfig()

	a := kingpin.New(filepath.Base(os.Args[0]), "Prometheus vCenter exporter")

	flag.AddFlags(a, &cfg.PromlogConfig)
	a.Version(version.Print("govc_exporter"))
	a.HelpFlag.Short('h')

	//Memory
	a.Flag("memlimit", "Memory (soft) limit in MB. Same as GOMEMLIMIT").Default("0").Int64Var(&cfg.MemoryLimitMB)

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
	a.Flag("scraper.vc.url", "vc api username").Envar("VC_URL").Required().StringVar(&cfg.ScraperConfig.VCenter)
	a.Flag("scraper.vc.username", "vc api username").Envar("VC_USERNAME").Required().StringVar(&cfg.ScraperConfig.Username)
	a.Flag("scraper.vc.password", "vc api password").Envar("VC_PASSWORD").Required().StringVar(&cfg.ScraperConfig.Password)
	a.Flag("scraper.client_pool_size", "number of simultanious requests to vCenter api").Default("5").IntVar(&cfg.ScraperConfig.ClientPoolSize)

	//scraper.cluster
	a.Flag("scraper.cluster", "Enable cluster sensor").Default("True").BoolVar(&cfg.ScraperConfig.Cluster.Enabled)
	a.Flag("scraper.cluster.max_age", "time in seconds clusters are cached").Default("5m").DurationVar(&cfg.ScraperConfig.Cluster.MaxAge)
	a.Flag("scraper.cluster.refresh_interval", "interval clusters are refreshed").Default("25s").DurationVar(&cfg.ScraperConfig.Cluster.RefreshInterval)

	//scraper.compute_resource
	a.Flag("scraper.compute_resource", "Enable compute_resource sensor").Default("True").BoolVar(&cfg.ScraperConfig.ComputeResource.Enabled)
	a.Flag("scraper.compute_resource.max_age", "time in seconds clusters are cached").Default("5m").DurationVar(&cfg.ScraperConfig.ComputeResource.MaxAge)
	a.Flag("scraper.compute_resource.refresh_interval", "interval clusters are refreshed").Default("25s").DurationVar(&cfg.ScraperConfig.ComputeResource.RefreshInterval)

	//scraper.datastore
	a.Flag("scraper.datastore", "Enable datastore sensor").Default("True").BoolVar(&cfg.ScraperConfig.Datastore.Enabled)
	a.Flag("scraper.datastore.max_age", "time in seconds datastores are cached").Default("2m").DurationVar(&cfg.ScraperConfig.Datastore.MaxAge)
	a.Flag("scraper.datastore.refresh_interval", "interval datastores are refreshed").Default("55s").DurationVar(&cfg.ScraperConfig.Datastore.RefreshInterval)

	//scraper.host
	a.Flag("scraper.host", "Enable host sensor").Default("True").BoolVar(&cfg.ScraperConfig.Host.Enabled)
	a.Flag("scraper.host.max_age", "time in seconds hosts are cached").Default("1m").DurationVar(&cfg.ScraperConfig.Host.MaxAge)
	a.Flag("scraper.host.refresh_interval", "interval hosts are refreshed").Default("25s").DurationVar(&cfg.ScraperConfig.Host.RefreshInterval)

	//scraper.host.perf
	a.Flag("scraper.host.perf", "Enable host performance metrics").Default("True").BoolVar(&cfg.ScraperConfig.HostPerf.Enabled)
	a.Flag("scraper.host.perf.max_age", "time in seconds performance metrics are cached").Default("10m").DurationVar(&cfg.ScraperConfig.HostPerf.MaxAge)
	a.Flag("scraper.host.perf.refresh_interval", "perf metrics refresh interval").Default("55s").DurationVar(&cfg.ScraperConfig.HostPerf.RefreshInterval)
	a.Flag("scraper.host.perf.max_sample_window", "max window metrics are collected").Default("5m").DurationVar(&cfg.ScraperConfig.HostPerf.MaxSampleWindow)
	a.Flag("scraper.host.perf.sample_interval", "time between metrics").Default("20s").DurationVar(&cfg.ScraperConfig.HostPerf.SampleInterval)
	a.Flag("scraper.host.perf.default_metrics", "Collect default host perf metrics").Default("True").BoolVar(&cfg.ScraperConfig.HostPerf.DefaultMetrics)
	a.Flag("scraper.host.perf.extra_metric", "Collect additional host perf metrics").StringsVar(&cfg.ScraperConfig.HostPerf.ExtraMetrics)

	//scraper.repool
	a.Flag("scraper.repool", "Enable resource pool sensor").Default("True").BoolVar(&cfg.ScraperConfig.ResourcePool.Enabled)
	a.Flag("scraper.repool.max_age", "time in seconds resource pools are cached").Default("2m").DurationVar(&cfg.ScraperConfig.ResourcePool.MaxAge)
	a.Flag("scraper.repool.refresh_interval", "interval resource pools are refreshed").Default("55s").DurationVar(&cfg.ScraperConfig.ResourcePool.RefreshInterval)

	//scraper.spod
	a.Flag("scraper.spod", "Enable datastore cluster sensor").Default("True").BoolVar(&cfg.ScraperConfig.Spod.Enabled)
	a.Flag("scraper.spod.max_age", "time in seconds spods are cached").Default("2m").DurationVar(&cfg.ScraperConfig.Spod.MaxAge)
	a.Flag("scraper.spod.refresh_interval", "interval spods are refreshed").Default("55s").DurationVar(&cfg.ScraperConfig.Spod.RefreshInterval)

	//scraper.tags
	a.Flag("scraper.tags", "Collect tags").Default("True").BoolVar(&cfg.ScraperConfig.Tags.Enabled)
	a.Flag("scraper.tags.max_age", "time in seconds tags are cached").Default("10m").DurationVar(&cfg.ScraperConfig.Tags.MaxAge)
	a.Flag("scraper.tags.refresh_interval", "interval tags are refreshed").Default("55s").DurationVar(&cfg.ScraperConfig.Tags.RefreshInterval)

	//scraper.vm
	a.Flag("scraper.vm", "Enable virtualmachine sensor").Default("True").BoolVar(&cfg.ScraperConfig.VirtualMachine.Enabled)
	a.Flag("scraper.vm.max_age", "time in seconds vm's are cached").Default("2m").DurationVar(&cfg.ScraperConfig.VirtualMachine.MaxAge)
	a.Flag("scraper.vm.refresh_interval", "interval vm's are refreshed").Default("55s").DurationVar(&cfg.ScraperConfig.VirtualMachine.RefreshInterval)
	a.Flag("scraper.vm.refresh_timeout", "the maximum amount of time a sensor refresh can take. Default is 3 times the refresh_interval").DurationVar(&cfg.ScraperConfig.VirtualMachine.RefreshTimeout)
	a.Flag("collector.vm.legacy", "Collect legacy metrics. Should all be available via scraper.vm.perf").Default("false").BoolVar(&cfg.CollectorConfig.VMLegacyMetrics)
	a.Flag("collector.vm.disk", "Collect extra vm disk metrics").Default("false").BoolVar(&cfg.CollectorConfig.VMAdvancedStorageMetrics)
	a.Flag("collector.vm.network", "Collect extra vm network metrics").Default("false").BoolVar(&cfg.CollectorConfig.VMAdvancedNetworkMetrics)

	// scraper.vm.perf
	a.Flag("scraper.vm.perf", "Enable vm performance metrics").Default("False").BoolVar(&cfg.ScraperConfig.VirtualMachinePerf.Enabled)
	a.Flag("scraper.vm.perf.max_age", "time in seconds perf metrics are cached").Default("10m").DurationVar(&cfg.ScraperConfig.VirtualMachinePerf.MaxAge)
	a.Flag("scraper.vm.perf.refresh_interval", "perf metrics refresh interval").Default("55s").DurationVar(&cfg.ScraperConfig.VirtualMachinePerf.RefreshInterval)
	a.Flag("scraper.vm.perf.refresh_timeout", "the maximum amount of time a sensor refresh can take. Default is 3 times the refresh_interval").DurationVar(&cfg.ScraperConfig.VirtualMachinePerf.RefreshTimeout)
	a.Flag("scraper.vm.perf.max_sample_window", "max window metrics are collected").Default("5m").DurationVar(&cfg.ScraperConfig.VirtualMachinePerf.MaxSampleWindow)
	a.Flag("scraper.vm.perf.sample_interval", "time between metrics").Default("20s").DurationVar(&cfg.ScraperConfig.VirtualMachinePerf.SampleInterval)
	a.Flag("scraper.vm.perf.default_metrics", "Collect default vm perf metrics").Default("True").BoolVar(&cfg.ScraperConfig.VirtualMachinePerf.DefaultMetrics)
	a.Flag("scraper.vm.perf.extra_metric", "Collect additional vm perf metrics").StringsVar(&cfg.ScraperConfig.VirtualMachinePerf.ExtraMetrics)

	// DB Backend
	a.Flag("scraper.backend.type", "type of backend").Default("memory").EnumVar(&cfg.ScraperConfig.Backend.Type, "memory", "redis")
	a.Flag("scraper.backend.redis.address", "Redis address").Default("localhost:6379").StringVar(&cfg.ScraperConfig.Backend.Redis.Address)
	a.Flag("scraper.backend.redis.password", "Redis password").Default("").StringVar(&cfg.ScraperConfig.Backend.Redis.Password)
	a.Flag("scraper.backend.redis.index", "Redis index").Default("0").IntVar(&cfg.ScraperConfig.Backend.Redis.Index)

	if _, err := a.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing commandline arguments: %v\n", err)
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	cfg.ScraperConfig.Tags.CategoryToCollect = helper.Union(
		cfg.CollectorConfig.ClusterTagLabels,
		cfg.CollectorConfig.DatastoreTagLabels,
		cfg.CollectorConfig.HostTagLabels,
		cfg.CollectorConfig.ResourcePoolTagLabels,
		cfg.CollectorConfig.StoragePodTagLabels,
		cfg.CollectorConfig.VMTagLabels,
	)

	return cfg
}
