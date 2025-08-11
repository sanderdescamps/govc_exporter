package config

import "fmt"

type CollectorConfig struct {
	UseIsecSpecifics       bool
	DisableExporterMetrics bool

	MaxRequests int

	ClusterTagLabels      []string
	DatastoreTagLabels    []string
	HostTagLabels         []string
	ResourcePoolTagLabels []string
	StoragePodTagLabels   []string

	VMLegacyMetrics          bool
	VMAdvancedNetworkMetrics bool
	VMAdvancedStorageMetrics bool
	VMTagLabels              []string

	HostStorageMetrics bool
}

func DefaultCollectorConf() CollectorConfig {
	return CollectorConfig{
		UseIsecSpecifics:       false,
		DisableExporterMetrics: false,

		MaxRequests: 10,

		ClusterTagLabels:      []string{},
		DatastoreTagLabels:    []string{},
		HostTagLabels:         []string{},
		ResourcePoolTagLabels: []string{},
		StoragePodTagLabels:   []string{},

		VMLegacyMetrics:          false,
		VMAdvancedNetworkMetrics: false,
		VMAdvancedStorageMetrics: false,
		VMTagLabels:              []string{},

		HostStorageMetrics: false,
	}
}

func (c CollectorConfig) Validate() error {
	if c.MaxRequests <= 0 {
		return fmt.Errorf("MaxRequests cannot be smaller than 1")
	}
	return nil
}
