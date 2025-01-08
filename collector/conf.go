package collector

type CollectorConf struct {
	CollectVMNetworks      bool
	CollectVMDisks         bool
	UseIsecSpecifics       bool
	DisableExporterMetrics bool
	MaxRequests            int
}
