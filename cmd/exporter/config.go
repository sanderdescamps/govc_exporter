package main

import (
	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/intrinsec/govc_exporter/collector"
	"github.com/intrinsec/govc_exporter/collector/scraper"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
)

type Config struct {
	ListenAddress   string
	MetricPath      string
	ScraperConfig   *scraper.ScraperConfig
	CollectorConfig *collector.CollectorConfig
	PromlogConfig   *promslog.Config
}

func LoadConfig() Config {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address on which to expose metrics and web interface.",
		).Default(":9752").String()
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		disableExporterMetrics = kingpin.Flag(
			"web.disable-exporter-metrics",
			"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
		).Bool()
		maxRequests = kingpin.Flag(
			"web.max-requests",
			"Maximum number of parallel scrape requests. Use 0 to disable.",
		).Default("40").Int()
		// disableDefaultCollectors = kingpin.Flag(
		// 	"collector.disable-defaults",
		// 	"Set all collectors to disabled by default.",
		// ).Default("false").Bool()
		// configFile = kingpin.Flag(
		// 	"web.config",
		// 	"[EXPERIMENTAL] Path to config yaml file that can enable TLS or authentication.",
		// ).Default("").String()

		endpoint = kingpin.Flag("collector.vc.url", "vc api username").Envar("VC_URL").Required().String()
		username = kingpin.Flag("collector.vc.username", "vc api username").Envar("VC_USERNAME").Required().String()
		password = kingpin.Flag("collector.vc.password", "vc api password").Envar("VC_PASSWORD").Required().String()

		// hostEnabled           = kingpin.Flag("scraper.esx", "Enable host metrics").Default("True").String() //Host always required
		datastoreEnabled      = kingpin.Flag("scraper.ds", "Enable datastore metrics").Default("True").String()
		repoolEnabled         = kingpin.Flag("scraper.repool", "Enable datastore metrics").Default("True").String()
		storagePodEnabled     = kingpin.Flag("scraper.spod", "Enable datastore metrics").Default("True").String()
		virtualMachineEnabled = kingpin.Flag("scraper.vm", "Enable datastore metrics").Default("True").String()
		clusterEnabled        = kingpin.Flag("scraper.cluster", "Enable datastore metrics").Default("True").String()

		hostMaxAgeSec                    = kingpin.Flag("scraper.host_max_age", "time in seconds host metrics are cached").Default("60").Int()
		hostRefreshIntervalSec           = kingpin.Flag("scraper.host_refresh_interval", "interval host metrics are refreshed").Default("25").Int()
		clusterMaxAgeSec                 = kingpin.Flag("scraper.cluster_max_age", "time in seconds cluster metrics are cached").Default("300").Int()
		clusterRefreshIntervalSec        = kingpin.Flag("scraper.cluster_refresh_interval", "interval cluster metrics are refreshed").Default("25").Int()
		virtualMachineMaxAgeSec          = kingpin.Flag("scraper.virtual_machine_max_age", "time in seconds vm metrics are cached").Default("120").Int()
		virtualMachineRefreshIntervalSec = kingpin.Flag("scraper.virtual_machine_refresh_interval", "interval vm metrics are refreshed").Default("55").Int()
		datastoreMaxAgeSec               = kingpin.Flag("scraper.datastore_max_age", "time in seconds datastore metrics are cached").Default("120").Int()
		datastoreRefreshIntervalSec      = kingpin.Flag("scraper.datastore_refresh_interval", "interval datastore metrics are refreshed").Default("55").Int()
		spodMaxAgeSec                    = kingpin.Flag("scraper.spod_max_age", "time in seconds spod metrics are cached").Default("120").Int()
		spodRefreshIntervalSec           = kingpin.Flag("scraper.storagepod_refresh_interval", "interval spod metrics are refreshed").Default("55").Int()
		resourcePoolMaxAgeSec            = kingpin.Flag("scraper.repool_max_age", "time in seconds resource pool metrics are cached").Default("120").Int()
		resourcePoolRefreshIntervalSec   = kingpin.Flag("scraper.repool_refresh_interval", "interval resource pool metrics are refreshed").Default("55").Int()
		onDemandCacheMaxAge              = kingpin.Flag("scraper.on_demand_cache_max_age", "time in seconds the scraper keeps all non-cache data. Used to get parent objects").Default("300").Int()
		cleanIntervalSec                 = kingpin.Flag("scraper.clean_interval", "interval the scraper cleans up old data").Default("5").Int()
		clientPoolSize                   = kingpin.Flag("scraper.client_pool_size", "number of simultanious requests to vCenter api").Default("5").Int()
		useIsecSpecifics                 = kingpin.Flag("collector.intrinsec", "Enable intrinsec specific features").Default("false").Bool()
		collectVMDisks                   = kingpin.Flag("collector.vm.disk", "Collect vm disk metrics").Default("false").Bool()
		collectVMNetwork                 = kingpin.Flag("collector.vm.network", "Collect vm network metrics").Default("false").Bool()
		CollectHostStorageMetrics        = kingpin.Flag("collector.host.storage", "Collect host storage metrics").Default("true").Bool()
	)

	promlogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("govc_exporter"))

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	return Config{
		ListenAddress: *listenAddress,
		MetricPath:    *metricsPath,
		PromlogConfig: promlogConfig,
		ScraperConfig: &scraper.ScraperConfig{
			Endpoint:                         *endpoint,
			Username:                         *username,
			Password:                         *password,
			HostMaxAgeSec:                    *hostMaxAgeSec,
			HostRefreshIntervalSec:           *hostRefreshIntervalSec,
			ClusterCollectorEnabled:          string2bool(*clusterEnabled),
			ClusterMaxAgeSec:                 *clusterMaxAgeSec,
			ClusterRefreshIntervalSec:        *clusterRefreshIntervalSec,
			VirtualMachineCollectorEnabled:   string2bool(*virtualMachineEnabled),
			VirtualMachineMaxAgeSec:          *virtualMachineMaxAgeSec,
			VirtualMachineRefreshIntervalSec: *virtualMachineRefreshIntervalSec,
			DatastoreCollectorEnabled:        string2bool(*datastoreEnabled),
			DatastoreMaxAgeSec:               *datastoreMaxAgeSec,
			DatastoreRefreshIntervalSec:      *datastoreRefreshIntervalSec,
			SpodCollectorEnabled:             string2bool(*storagePodEnabled),
			SpodMaxAgeSec:                    *spodMaxAgeSec,
			SpodRefreshIntervalSec:           *spodRefreshIntervalSec,
			ResourcePoolCollectorEnabled:     string2bool(*repoolEnabled),
			ResourcePoolMaxAgeSec:            *resourcePoolMaxAgeSec,
			ResourcePoolRefreshIntervalSec:   *resourcePoolRefreshIntervalSec,
			OnDemandCacheMaxAge:              *onDemandCacheMaxAge,
			CleanIntervalSec:                 *cleanIntervalSec,
			ClientPoolSize:                   *clientPoolSize,
		},
		CollectorConfig: &collector.CollectorConfig{
			UseIsecSpecifics:          *useIsecSpecifics,
			CollectVMNetworks:         *collectVMNetwork,
			CollectVMDisks:            *collectVMDisks,
			CollectHostStorageMetrics: *CollectHostStorageMetrics,
			DisableExporterMetrics:    *disableExporterMetrics,
			MaxRequests:               *maxRequests,
		},
	}
}
