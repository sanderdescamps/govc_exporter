package scraper

import (
	"fmt"
	"net/url"

	"github.com/vmware/govmomi/vim25/soap"
)

type ScraperConfig struct {
	Endpoint                         string `json:"endpoint"`
	Username                         string `json:"username"`
	Password                         string `json:"password"`
	HostCollectorEnabled             bool   `json:"host_collector_enabled"`
	HostMaxAgeSec                    int    `json:"host_max_age_sec"`
	HostRefreshIntervalSec           int    `json:"host_refresh_interval_sec"`
	ClusterCollectorEnabled          bool   `json:"cluster_collector_enabled"`
	ClusterMaxAgeSec                 int    `json:"cluster_max_age_sec"`
	ClusterRefreshIntervalSec        int    `json:"cluster_refresh_interval_sec"`
	VirtualMachineCollectorEnabled   bool   `json:"virtual_machine_collector_enabled"`
	VirtualMachineMaxAgeSec          int    `json:"virtual_machine_max_age_sec"`
	VirtualMachineRefreshIntervalSec int    `json:"virtual_machine_refresh_interval_sec"`
	DatastoreCollectorEnabled        bool   `json:"datastore_collector_enabled"`
	DatastoreMaxAgeSec               int    `json:"datastore_max_age_sec"`
	DatastoreRefreshIntervalSec      int    `json:"datastore_refresh_interval_sec"`
	SpodCollectorEnabled             bool   `json:"storagepod_collector_enabled"`
	SpodMaxAgeSec                    int    `json:"storagepod_max_age_sec"`
	SpodRefreshIntervalSec           int    `json:"storagepod_refresh_interval_sec"`
	ResourcePoolCollectorEnabled     bool   `json:"resource_pool_collector_enabled"`
	ResourcePoolMaxAgeSec            int    `json:"resource_pool_max_age_sec"`
	ResourcePoolRefreshIntervalSec   int    `json:"resource_pool_refresh_interval_sec"`

	OnDemandCacheMaxAge int `json:"on_demand_cache_max_age_sec"`
	CleanIntervalSec    int `json:"clean_interval_sec"`
	ClientPoolSize      int `json:"client_pool_size"`
}

func NewDefaultScraperConfig() ScraperConfig {
	return ScraperConfig{
		HostCollectorEnabled:             true,
		HostMaxAgeSec:                    120,
		HostRefreshIntervalSec:           60,
		ClusterCollectorEnabled:          true,
		ClusterMaxAgeSec:                 300,
		ClusterRefreshIntervalSec:        30,
		VirtualMachineCollectorEnabled:   true,
		VirtualMachineMaxAgeSec:          120,
		VirtualMachineRefreshIntervalSec: 60,
		DatastoreCollectorEnabled:        true,
		DatastoreMaxAgeSec:               120,
		DatastoreRefreshIntervalSec:      30,
		SpodCollectorEnabled:             true,
		SpodMaxAgeSec:                    120,
		SpodRefreshIntervalSec:           60,
		ResourcePoolCollectorEnabled:     true,
		ResourcePoolMaxAgeSec:            120,
		ResourcePoolRefreshIntervalSec:   60,
		OnDemandCacheMaxAge:              300,
		CleanIntervalSec:                 5,
		ClientPoolSize:                   5,
	}
}

func (c ScraperConfig) URL() (*url.URL, error) {
	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url %s", c.Endpoint)
	}

	port := u.Port()
	if u.Port() == "" {
		if u.Scheme == "https" {
			port = "443"
		} else if u.Scheme == "http" {
			port = "80"
		}
	}

	urlToParse := fmt.Sprintf("%s://%s", u.Scheme, u.Hostname())
	if port != "" {
		urlToParse = fmt.Sprintf("%s:%s", urlToParse, port)
	}
	if u.Path != "" {
		urlToParse = fmt.Sprintf("%s%s", urlToParse, u.Path)
	}
	if len(u.Query()) > 0 {
		urlToParse = fmt.Sprintf("%s?%s", urlToParse, u.Query().Encode())
	}
	if u.Fragment != "" {
		urlToParse = fmt.Sprintf("%s#%s", urlToParse, u.Fragment)
	}

	parseURL, err := url.Parse(urlToParse)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url  %s", urlToParse)
	}
	return parseURL, nil
}

func (c ScraperConfig) SoapURL() (*url.URL, error) {
	u, err := soap.ParseURL(c.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url %s", c.Endpoint)
	}
	return u, err
}
