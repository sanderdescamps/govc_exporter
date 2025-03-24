package main

import (
	"time"

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
	ScraperConfig      *scraper.ScraperConfig
	CollectorConfig    *collector.CollectorConfig
	PromlogConfig      *promslog.Config
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
		allowManualRefresh = kingpin.Flag(
			"web.manual-refresh",
			"Enable /refresh/{sensor} path to trigger a refresh of a sensor.",
		).Default("false").Bool()
		allowDumps = kingpin.Flag(
			"web.allow-dumps",
			"Enable /dump path to trigger a dump of the cache data in ./dumps folder on server side. Only enable for debugging.",
		).Default("false").Bool()
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

		hostEnabled            = kingpin.Flag("scraper.esx", "Enable host scraper").Default("True").Bool()
		datastoreEnabled       = kingpin.Flag("scraper.ds", "Enable datastore scraper").Default("True").Bool()
		repoolEnabled          = kingpin.Flag("scraper.repool", "Enable resource pool scraper").Default("True").Bool()
		storagePodEnabled      = kingpin.Flag("scraper.spod", "Enable datastore cluster scraper").Default("True").Bool()
		virtualMachineEnabled  = kingpin.Flag("scraper.vm", "Enable virtualmachine scraper").Default("True").Bool()
		computeResourceEnabled = kingpin.Flag("scraper.cluster", "Enable cluster scraper").Default("True").Bool()
		clusterEnabled         = kingpin.Flag("scraper.compute_resource", "Enable compute_resource scraper").Default("True").Bool()
		tagsEnabled            = kingpin.Flag("scraper.tags", "Collect tags").Default("True").Bool()

		hostMaxAgeSec                     = kingpin.Flag("scraper.host.max_age", "time in seconds hosts are cached").Default("60").Int64()
		hostRefreshIntervalSec            = kingpin.Flag("scraper.host.refresh_interval", "interval hosts are refreshed").Default("25").Int64()
		computeResourceMaxAgeSec          = kingpin.Flag("scraper.compute_resource.max_age", "time in seconds clusters are cached").Default("300").Int64()
		computeResourceRefreshIntervalSec = kingpin.Flag("scraper.compute_resource.refresh_interval", "interval clusters are refreshed").Default("25").Int64()
		clusterMaxAgeSec                  = kingpin.Flag("scraper.cluster.max_age", "time in seconds clusters are cached").Default("300").Int64()
		clusterRefreshIntervalSec         = kingpin.Flag("scraper.cluster.refresh_interval", "interval clusters are refreshed").Default("25").Int64()
		virtualMachineMaxAgeSec           = kingpin.Flag("scraper.vm.max_age", "time in seconds vm's are cached").Default("120").Int64()
		virtualMachineRefreshIntervalSec  = kingpin.Flag("scraper.vm.refresh_interval", "interval vm's are refreshed").Default("55").Int64()
		datastoreMaxAgeSec                = kingpin.Flag("scraper.datastore.max_age", "time in seconds datastores are cached").Default("120").Int64()
		datastoreRefreshIntervalSec       = kingpin.Flag("scraper.datastore.refresh_interval", "interval datastores are refreshed").Default("55").Int64()
		spodMaxAgeSec                     = kingpin.Flag("scraper.spod.max_age", "time in seconds spods are cached").Default("120").Int64()
		spodRefreshIntervalSec            = kingpin.Flag("scraper.spod.refresh_interval", "interval spods are refreshed").Default("55").Int64()
		resourcePoolMaxAgeSec             = kingpin.Flag("scraper.repool.max_age", "time in seconds resource pools are cached").Default("120").Int64()
		resourcePoolRefreshIntervalSec    = kingpin.Flag("scraper.repool.refresh_interval", "interval resource pools are refreshed").Default("55").Int64()
		tagsMaxAgeSec                     = kingpin.Flag("scraper.tags.max_age", "time in seconds tags are cached").Default("600").Int64()
		tagsRefreshIntervalSec            = kingpin.Flag("scraper.tags.refresh_interval", "interval tags are refreshed").Default("290").Int64()

		onDemandCacheMaxAge           = kingpin.Flag("scraper.on_demand_cache.max_age", "time in seconds the scraper keeps all non-cache data. Used to get parent objects").Default("300").Int64()
		cleanIntervalSec              = kingpin.Flag("scraper.clean_interval", "interval the scraper cleans up old data").Default("5").Int64()
		clientPoolSize                = kingpin.Flag("scraper.client_pool_size", "number of simultanious requests to vCenter api").Default("5").Int()
		useIsecSpecifics              = kingpin.Flag("collector.intrinsec", "Enable intrinsec specific features").Default("false").Bool()
		virtualMachineAdvancedStorage = kingpin.Flag("collector.vm.disk", "Collect extra vm disk metrics").Default("false").Bool()
		virtualMachineAdvancedNetwork = kingpin.Flag("collector.vm.network", "Collect extra vm network metrics").Default("false").Bool()

		hostStorage = kingpin.Flag("collector.host.storage", "Collect host storage metrics").Default("false").Bool()

		clusterTagLabel        = kingpin.Flag("collector.cluster.tag_label", "List of vmware tag categories to collect which will be added as label in metrics").Strings()
		datastoreTagLabel      = kingpin.Flag("collector.datastore.tag_label", "List of vmware tag categories to collect which will be added as label in metrics").Strings()
		hostTagLabel           = kingpin.Flag("collector.host.tag_label", "List of vmware tag categories which will be added as label in metrics").Strings()
		resourcePoolTagLabel   = kingpin.Flag("collector.repool.tag_label", "List of tag categories which will be added as label in metrics").Strings()
		storagePodTagLabel     = kingpin.Flag("collector.spod.tag_label", "List of vmware tag categories to collect which will be added as label in metrics").Strings()
		virtualMachineTagLabel = kingpin.Flag("collector.vm.tag_label", "List of vmware tag categories to collect which will be added as label in metrics").Strings()
	)

	promlogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("govc_exporter"))

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	scraperLabelCats := mergeLists(
		virtualMachineTagLabel,
		hostTagLabel,
		datastoreTagLabel,
		storagePodTagLabel,
		resourcePoolTagLabel,
		clusterTagLabel,
	)

	return Config{
		ListenAddress:      *listenAddress,
		MetricPath:         *metricsPath,
		PromlogConfig:      promlogConfig,
		AllowDumps:         *allowDumps,
		AllowManualRefresh: *allowManualRefresh,
		ScraperConfig: &scraper.ScraperConfig{
			Endpoint:                       *endpoint,
			Username:                       *username,
			Password:                       *password,
			HostScraperEnabled:             *hostEnabled,
			HostMaxAge:                     time.Duration(*hostMaxAgeSec) * time.Second,
			HostRefreshInterval:            time.Duration(*hostRefreshIntervalSec) * time.Second,
			ComputeResourceScraperEnabled:  *computeResourceEnabled,
			ComputeResourceMaxAge:          time.Duration(*computeResourceMaxAgeSec) * time.Second,
			ComputeResourceRefreshInterval: time.Duration(*computeResourceRefreshIntervalSec) * time.Second,
			ClusterScraperEnabled:          *clusterEnabled,
			ClusterMaxAge:                  time.Duration(*clusterMaxAgeSec) * time.Second,
			ClusterRefreshInterval:         time.Duration(*clusterRefreshIntervalSec) * time.Second,
			VirtualMachineScraperEnabled:   *virtualMachineEnabled,
			VirtualMachineMaxAge:           time.Duration(*virtualMachineMaxAgeSec) * time.Second,
			VirtualMachineRefreshInterval:  time.Duration(*virtualMachineRefreshIntervalSec) * time.Second,
			DatastoreScraperEnabled:        *datastoreEnabled,
			DatastoreMaxAge:                time.Duration(*datastoreMaxAgeSec) * time.Second,
			DatastoreRefreshInterval:       time.Duration(*datastoreRefreshIntervalSec) * time.Second,
			SpodScraperEnabled:             *storagePodEnabled,
			SpodMaxAge:                     time.Duration(*spodMaxAgeSec) * time.Second,
			SpodRefreshInterval:            time.Duration(*spodRefreshIntervalSec) * time.Second,
			ResourcePoolScraperEnabled:     *repoolEnabled,
			ResourcePoolMaxAge:             time.Duration(*resourcePoolMaxAgeSec) * time.Second,
			ResourcePoolRefreshInterval:    time.Duration(*resourcePoolRefreshIntervalSec) * time.Second,
			TagsScraperEnabled:             *tagsEnabled,
			TagsCategoryToCollect:          scraperLabelCats,
			TagsMaxAge:                     time.Duration(*tagsMaxAgeSec) * time.Second,
			TagsRefreshInterval:            time.Duration(*tagsRefreshIntervalSec) * time.Second,

			OnDemandCacheMaxAge: time.Duration(*onDemandCacheMaxAge) * time.Second,
			CleanInterval:       time.Duration(*cleanIntervalSec) * time.Second,
			ClientPoolSize:      *clientPoolSize,
		},
		CollectorConfig: &collector.CollectorConfig{
			UseIsecSpecifics:         *useIsecSpecifics,
			VMAdvancedNetworkMetrics: *virtualMachineAdvancedNetwork,
			VMAdvancedStorageMetrics: *virtualMachineAdvancedStorage,
			HostStorageMetrics:       *hostStorage,
			DisableExporterMetrics:   *disableExporterMetrics,
			MaxRequests:              *maxRequests,

			ClusterTagLabels:      mergeLists(clusterTagLabel),
			DatastoreTagLabels:    mergeLists(datastoreTagLabel),
			HostTagLabels:         mergeLists(hostTagLabel),
			ResourcePoolTagLabels: mergeLists(resourcePoolTagLabel),
			StoragePodTagLabels:   mergeLists(storagePodTagLabel),
			VMTagLabels:           mergeLists(virtualMachineTagLabel),
		},
	}
}
