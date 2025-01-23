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
		tagsEnabled           = kingpin.Flag("scraper.tags", "Collect tags").Default("True").String()

		hostMaxAgeSec                    = kingpin.Flag("scraper.host.max_age", "time in seconds hosts are cached").Default("60").Int64()
		hostRefreshIntervalSec           = kingpin.Flag("scraper.host.refresh_interval", "interval hosts are refreshed").Default("25").Int64()
		clusterMaxAgeSec                 = kingpin.Flag("scraper.cluster.max_age", "time in seconds clusters are cached").Default("300").Int64()
		clusterRefreshIntervalSec        = kingpin.Flag("scraper.cluster.refresh_interval", "interval clusters are refreshed").Default("25").Int64()
		virtualMachineMaxAgeSec          = kingpin.Flag("scraper.vm.max_age", "time in seconds vm's are cached").Default("120").Int64()
		virtualMachineRefreshIntervalSec = kingpin.Flag("scraper.vm.refresh_interval", "interval vm's are refreshed").Default("55").Int64()
		datastoreMaxAgeSec               = kingpin.Flag("scraper.datastore.max_age", "time in seconds datastores are cached").Default("120").Int64()
		datastoreRefreshIntervalSec      = kingpin.Flag("scraper.datastore.refresh_interval", "interval datastores are refreshed").Default("55").Int64()
		spodMaxAgeSec                    = kingpin.Flag("scraper.spod.max_age", "time in seconds spods are cached").Default("120").Int64()
		spodRefreshIntervalSec           = kingpin.Flag("scraper.spod.refresh_interval", "interval spods are refreshed").Default("55").Int64()
		resourcePoolMaxAgeSec            = kingpin.Flag("scraper.repool.max_age", "time in seconds resource pools are cached").Default("120").Int64()
		resourcePoolRefreshIntervalSec   = kingpin.Flag("scraper.repool.refresh_interval", "interval resource pools are refreshed").Default("55").Int64()
		tagsMaxAgeSec                    = kingpin.Flag("scraper.tags.max_age", "time in seconds tags are cached").Default("600").Int64()
		tagsRefreshIntervalSec           = kingpin.Flag("scraper.tags.refresh_interval", "interval tags are refreshed").Default("290").Int64()

		onDemandCacheMaxAge           = kingpin.Flag("scraper.on_demand_cache.max_age", "time in seconds the scraper keeps all non-cache data. Used to get parent objects").Default("300").Int64()
		cleanIntervalSec              = kingpin.Flag("scraper.clean_interval", "interval the scraper cleans up old data").Default("5").Int64()
		clientPoolSize                = kingpin.Flag("scraper.client_pool_size", "number of simultanious requests to vCenter api").Default("5").Int()
		useIsecSpecifics              = kingpin.Flag("collector.intrinsec", "Enable intrinsec specific features").Default("false").Bool()
		virtualMachineAdvancedStorage = kingpin.Flag("collector.vm.disk", "Collect extra vm disk metrics").Default("false").Bool()
		virtualMachineAdvancedNetwork = kingpin.Flag("collector.vm.network", "Collect extra vm network metrics").Default("false").Bool()

		hostStorage = kingpin.Flag("collector.host.storage", "Collect host storage metrics").Default("true").Bool()

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
			TagsCollectorEnbled:              string2bool(*tagsEnabled),
			TagsCategoryToCollect:            scraperLabelCats,
			TagsMaxAgeSec:                    *tagsMaxAgeSec,
			TagsRefreshIntervalSec:           *tagsRefreshIntervalSec,

			OnDemandCacheMaxAge: *onDemandCacheMaxAge,
			CleanIntervalSec:    *cleanIntervalSec,
			ClientPoolSize:      *clientPoolSize,
		},
		CollectorConfig: &collector.CollectorConfig{
			UseIsecSpecifics:          *useIsecSpecifics,
			VMAdvancedNetworkMetrics:  *virtualMachineAdvancedNetwork,
			VMAdvancedStorageMetrics:  *virtualMachineAdvancedStorage,
			CollectHostStorageMetrics: *hostStorage,
			DisableExporterMetrics:    *disableExporterMetrics,
			MaxRequests:               *maxRequests,

			ClusterTagLabels:      mergeLists(clusterTagLabel),
			DatastoreTagLabels:    mergeLists(datastoreTagLabel),
			HostTagLabels:         mergeLists(hostTagLabel),
			ResourcePoolTagLabels: mergeLists(resourcePoolTagLabel),
			StoragePodTagLabels:   mergeLists(storagePodTagLabel),
			VMTagLabels:           mergeLists(virtualMachineTagLabel),
		},
	}
}
