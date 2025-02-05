package scraper

import (
	"fmt"
	"net/url"
	"time"

	"github.com/vmware/govmomi/vim25/soap"
)

type ScraperConfig struct {
	Endpoint                       string
	Username                       string
	Password                       string
	HostScraperEnabled             bool
	HostMaxAge                     time.Duration
	HostRefreshInterval            time.Duration
	ClusterScraperEnabled          bool
	ClusterMaxAge                  time.Duration
	ClusterRefreshInterval         time.Duration
	ComputeResourceScraperEnabled  bool
	ComputeResourceMaxAge          time.Duration
	ComputeResourceRefreshInterval time.Duration
	VirtualMachineScraperEnabled   bool
	VirtualMachineMaxAge           time.Duration
	VirtualMachineRefreshInterval  time.Duration
	DatastoreScraperEnabled        bool
	DatastoreMaxAge                time.Duration
	DatastoreRefreshInterval       time.Duration
	SpodScraperEnabled             bool
	SpodMaxAge                     time.Duration
	SpodRefreshInterval            time.Duration
	ResourcePoolScraperEnabled     bool
	ResourcePoolMaxAge             time.Duration
	ResourcePoolRefreshInterval    time.Duration
	TagsScraperEnabled             bool
	TagsCategoryToCollect          []string
	TagsMaxAge                     time.Duration
	TagsRefreshInterval            time.Duration

	OnDemandCacheMaxAge time.Duration
	CleanInterval       time.Duration
	ClientPoolSize      int
}

func NewDefaultScraperConfig() ScraperConfig {
	return ScraperConfig{
		HostScraperEnabled:            true,
		HostMaxAge:                    time.Duration(120) * time.Second,
		HostRefreshInterval:           time.Duration(60) * time.Second,
		ClusterScraperEnabled:         true,
		ClusterMaxAge:                 time.Duration(300) * time.Second,
		ClusterRefreshInterval:        time.Duration(30) * time.Second,
		VirtualMachineScraperEnabled:  true,
		VirtualMachineMaxAge:          time.Duration(120) * time.Second,
		VirtualMachineRefreshInterval: time.Duration(60) * time.Second,
		DatastoreScraperEnabled:       true,
		DatastoreMaxAge:               time.Duration(120) * time.Second,
		DatastoreRefreshInterval:      time.Duration(30) * time.Second,
		SpodScraperEnabled:            true,
		SpodMaxAge:                    time.Duration(120) * time.Second,
		SpodRefreshInterval:           time.Duration(60) * time.Second,
		ResourcePoolScraperEnabled:    true,
		ResourcePoolMaxAge:            time.Duration(120) * time.Second,
		ResourcePoolRefreshInterval:   time.Duration(60) * time.Second,
		TagsScraperEnabled:            true,
		TagsCategoryToCollect:         []string{},
		TagsMaxAge:                    time.Duration(600) * time.Second,
		TagsRefreshInterval:           time.Duration(290) * time.Second,
		OnDemandCacheMaxAge:           300,
		CleanInterval:                 time.Duration(5) * time.Second,
		ClientPoolSize:                5,
	}
}

func (c ScraperConfig) Validate() error {
	if !c.HostScraperEnabled && c.VirtualMachineScraperEnabled {
		return fmt.Errorf(`HostScraperEnabled must be enabled when 
VirtualMachineScraperEnabled is enabled because scraper needs the hosts 
when it queries the vm's`)
	}

	if c.ClusterMaxAge.Seconds()+5 <= c.ClusterRefreshInterval.Seconds() {
		return fmt.Errorf("ClusterMaxAge must be more than 5sec bigger than ClusterRefreshInterval")
	}
	if c.ComputeResourceMaxAge.Seconds()+5 <= c.ComputeResourceRefreshInterval.Seconds() {
		return fmt.Errorf("ComputeResourceMaxAge must be more than 5sec bigger than ComputeResourceRefreshInterval")
	}
	if c.DatastoreMaxAge.Seconds()+5 <= c.DatastoreRefreshInterval.Seconds() {
		return fmt.Errorf("DatastoreMaxAge must be more than 5sec bigger than DatastoreRefreshInterval")
	}
	if c.HostMaxAge.Seconds()+5 <= c.HostRefreshInterval.Seconds() {
		return fmt.Errorf("HostMaxAge must be more than 5sec bigger than HostRefreshInterval")
	}
	if c.ResourcePoolMaxAge.Seconds()+5 <= c.ResourcePoolRefreshInterval.Seconds() {
		return fmt.Errorf("ResourcePoolMaxAge must be more than 5sec bigger than ResourcePoolRefreshInterval")
	}
	if c.SpodMaxAge.Seconds()+5 <= c.SpodRefreshInterval.Seconds() {
		return fmt.Errorf("SpodMaxAge must be more than 5sec bigger than SpodRefreshInterval")
	}
	if c.TagsMaxAge.Seconds()+5 <= c.TagsRefreshInterval.Seconds() {
		return fmt.Errorf("TagsMaxAge must be more than 5sec bigger than TagsRefreshInterval")
	}
	if c.VirtualMachineMaxAge.Seconds()+5 <= c.VirtualMachineRefreshInterval.Seconds() {
		return fmt.Errorf("VirtualMachineMaxAge must be more than 5sec bigger than VirtualMachineRefreshInterval")
	}
	return nil
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
