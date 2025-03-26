package scraper

import (
	"fmt"
	"net/url"
	"time"

	"github.com/vmware/govmomi/vim25/soap"
)

type Config struct {
	Endpoint           string
	Username           string
	Password           string
	Cluster            SensorConfig
	ComputeResource    SensorConfig
	Datastore          SensorConfig
	Host               SensorConfig
	HostPerf           PerfSensorConfig
	ResourcePool       SensorConfig
	Spod               SensorConfig
	Tags               TagsSensorConfig
	VirtualMachine     SensorConfig
	VirtualMachinePerf PerfSensorConfig
	OnDemand           struct {
		MaxAge        time.Duration
		CleanInterval time.Duration
	}
	// CleanInterval  time.Duration
	ClientPoolSize int
}

type SensorConfig struct {
	Enabled         bool
	MaxAge          time.Duration
	RefreshInterval time.Duration
	CleanInterval   time.Duration
}

type PerfSensorConfig struct {
	Enabled         bool
	MaxAge          time.Duration
	RefreshInterval time.Duration
	CleanInterval   time.Duration
	MaxSampleWindow time.Duration
	SampleInterval  time.Duration
	DefaultMetrics  bool
	ExtraMetrics    []string
}

func (c *PerfSensorConfig) SensorConfig() SensorConfig {
	return SensorConfig{
		Enabled:         c.Enabled,
		MaxAge:          c.MaxAge,
		RefreshInterval: c.RefreshInterval,
		CleanInterval:   c.CleanInterval,
	}
}

type TagsSensorConfig struct {
	SensorConfig
	CategoryToCollect []string
}

func DefaultConfig() Config {
	return Config{
		Cluster: SensorConfig{
			Enabled:         true,
			MaxAge:          time.Duration(120) * time.Second,
			RefreshInterval: time.Duration(60) * time.Second,
		},
		ComputeResource: SensorConfig{
			Enabled:         true,
			MaxAge:          time.Duration(120) * time.Second,
			RefreshInterval: time.Duration(60) * time.Second,
		},
		Datastore: SensorConfig{
			Enabled:         true,
			MaxAge:          time.Duration(120) * time.Second,
			RefreshInterval: time.Duration(30) * time.Second,
		},
		Host: SensorConfig{
			Enabled:         true,
			MaxAge:          time.Duration(120) * time.Second,
			RefreshInterval: time.Duration(30) * time.Second,
		},
		ResourcePool: SensorConfig{
			Enabled:         true,
			MaxAge:          time.Duration(120) * time.Second,
			RefreshInterval: time.Duration(60) * time.Second,
		},
		Spod: SensorConfig{
			Enabled:         true,
			MaxAge:          time.Duration(120) * time.Second,
			RefreshInterval: time.Duration(60) * time.Second,
		},
		Tags: TagsSensorConfig{
			SensorConfig: SensorConfig{
				Enabled:         true,
				MaxAge:          time.Duration(600) * time.Second,
				RefreshInterval: time.Duration(290) * time.Second,
			},
			CategoryToCollect: []string{},
		},
		VirtualMachine: SensorConfig{
			Enabled:         true,
			MaxAge:          time.Duration(120) * time.Second,
			RefreshInterval: time.Duration(60) * time.Second,
		},

		OnDemand: struct {
			MaxAge        time.Duration
			CleanInterval time.Duration
		}{
			MaxAge:        300,
			CleanInterval: time.Duration(5) * time.Second,
		},
		ClientPoolSize: 5,
	}
}

func (c Config) Validate() error {
	if !c.Host.Enabled && c.VirtualMachine.Enabled {
		return fmt.Errorf(`HostScraperEnabled must be enabled when 
VirtualMachineScraperEnabled is enabled because scraper needs the hosts 
when it queries the vm's`)
	}

	if c.Cluster.MaxAge.Seconds()+5 <= c.Cluster.RefreshInterval.Seconds() {
		return fmt.Errorf("ClusterMaxAge must be more than 5sec bigger than ClusterRefreshInterval")
	}
	if c.ComputeResource.MaxAge.Seconds()+5 <= c.ComputeResource.RefreshInterval.Seconds() {
		return fmt.Errorf("ComputeResourceMaxAge must be more than 5sec bigger than ComputeResourceRefreshInterval")
	}
	if c.Datastore.MaxAge.Seconds()+5 <= c.Datastore.RefreshInterval.Seconds() {
		return fmt.Errorf("DatastoreMaxAge must be more than 5sec bigger than DatastoreRefreshInterval")
	}
	if c.Host.MaxAge.Seconds()+5 <= c.Host.RefreshInterval.Seconds() {
		return fmt.Errorf("HostMaxAge must be more than 5sec bigger than HostRefreshInterval")
	}
	if c.ResourcePool.MaxAge.Seconds()+5 <= c.ResourcePool.RefreshInterval.Seconds() {
		return fmt.Errorf("ResourcePoolMaxAge must be more than 5sec bigger than ResourcePoolRefreshInterval")
	}
	if c.Spod.MaxAge.Seconds()+5 <= c.Spod.RefreshInterval.Seconds() {
		return fmt.Errorf("SpodMaxAge must be more than 5sec bigger than SpodRefreshInterval")
	}
	if c.Tags.MaxAge.Seconds()+5 <= c.Tags.RefreshInterval.Seconds() {
		return fmt.Errorf("TagsMaxAge must be more than 5sec bigger than TagsRefreshInterval")
	}
	if c.VirtualMachine.MaxAge.Seconds()+5 <= c.VirtualMachine.RefreshInterval.Seconds() {
		return fmt.Errorf("VirtualMachineMaxAge must be more than 5sec bigger than VirtualMachineRefreshInterval")
	}
	return nil
}

func (c Config) URL() (*url.URL, error) {
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

func (c Config) SoapURL() (*url.URL, error) {
	u, err := soap.ParseURL(c.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url %s", c.Endpoint)
	}
	return u, err
}
