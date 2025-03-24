package collector

import "fmt"

type Config struct {
	UseIsecSpecifics       bool
	DisableExporterMetrics bool

	MaxRequests int

	ClusterTagLabels      []string
	DatastoreTagLabels    []string
	HostTagLabels         []string
	ResourcePoolTagLabels []string
	StoragePodTagLabels   []string

	VMAdvancedNetworkMetrics bool
	VMAdvancedStorageMetrics bool
	VMTagLabels              []string

	HostStorageMetrics bool
}

func DefaultCollectorConf() *Config {
	return &Config{
		UseIsecSpecifics:       false,
		DisableExporterMetrics: false,

		MaxRequests: 10,

		ClusterTagLabels:      []string{},
		DatastoreTagLabels:    []string{},
		HostTagLabels:         []string{},
		ResourcePoolTagLabels: []string{},
		StoragePodTagLabels:   []string{},

		VMAdvancedNetworkMetrics: false,
		VMAdvancedStorageMetrics: false,
		VMTagLabels:              []string{},

		HostStorageMetrics: false,
	}
}

func (c Config) Validate() error {
	if c.MaxRequests <= 0 {
		return fmt.Errorf("MaxRequests cannot be smaller than 1")
	}
	return nil
}
