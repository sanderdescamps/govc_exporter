package config

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/vim25/soap"
)

const (
	URL_REGEX_PATTERN = `(?:(https)?:\/\/)?([-a-zA-Z0-9%._\+~]{1,256})(?:\:(\d+))?(.*)\/?`
)

type ScraperConfig struct {
	VCenter  string
	Username string
	Password string

	Backend BackendConfig

	Cluster            SensorConfig
	ComputeResource    SensorConfig
	Datastore          SensorConfig
	Datacenter         SensorConfig
	Folder             SensorConfig
	Host               SensorConfig
	HostPerf           PerfSensorConfig
	ResourcePool       SensorConfig
	Spod               SensorConfig
	Tags               TagsSensorConfig
	VirtualMachine     SensorConfig
	VirtualMachinePerf PerfSensorConfig
	// CleanInterval  time.Duration
	ClientPoolSize int
}

type SensorConfig struct {
	Enabled         bool
	MaxAge          time.Duration
	RefreshInterval time.Duration
	RefreshTimeout  time.Duration
}

type PerfSensorConfig struct {
	Enabled         bool
	MaxAge          time.Duration
	RefreshInterval time.Duration
	RefreshTimeout  time.Duration
	MaxSampleWindow time.Duration
	SampleInterval  time.Duration
	DefaultMetrics  bool
	ExtraMetrics    []string
	Filters         []string
}

type PerfFilter struct {
	MatchName     regexp.Regexp
	MatchInstance regexp.Regexp
	Action        PerfFilterAction
	NewName       string
}

type PerfFilterAction string

const (
	PerfFilterActionDrop           = PerfFilterAction("drop")
	PerfFilterActionSum            = PerfFilterAction("sum")
	PerfFilterActionSpit           = PerfFilterAction("split") //Special mode that generates exrta mertics for testing
	PerfFilterActionRename         = PerfFilterAction("rename")
	PerfFilterActionInstanceRename = PerfFilterAction("instance-rename")
)

func (c PerfSensorConfig) ParseFilters() ([]PerfFilter, error) {
	filterStrings := c.Filters
	if helper.Contains(filterStrings, "vcsim-fix") {
		filterStrings = helper.Remove(filterStrings, "vcsim-fix")
		filterStrings = slices.Concat([]string{
			"*;\\*;instance-rename;",
			"cpu\\.usage\\.average;;split;0,1,2,3,",
			"net\\.bytes(Rx|Tx)\\.average;;split;vmnic0,vmnic1,vmnic2,vmnic3,",
			"net\\.(errors|dropped)(Rx|Tx)\\.summation;;split;vmnic0,vmnic1,vmnic2,vmnic3,",
		}, filterStrings)
	}

	pFilters := []PerfFilter{}
	for _, f := range filterStrings {
		var pFilter PerfFilter

		s := strings.Split(f, ";")
		if l := len(s); l < 1 {
			return nil, fmt.Errorf("invalid filter string")
		}

		if len(s) >= 1 {
			switch pattern := s[0]; pattern {
			case "*":
				pFilter.MatchName = *regexp.MustCompile(".*")
			case "":
				pFilter.MatchName = *regexp.MustCompile("^$")
			default:
				m, err := regexp.Compile(pattern)
				if err != nil {
					return nil, fmt.Errorf("invalid filter string")
				}
				pFilter.MatchName = *m
			}
		}

		if len(s) >= 2 {
			switch pattern := s[1]; pattern {
			case "*":
				pFilter.MatchInstance = *regexp.MustCompile(".*")
			case "":
				pFilter.MatchInstance = *regexp.MustCompile("^$")
			default:
				m, err := regexp.Compile(pattern)
				if err != nil {
					return nil, fmt.Errorf("invalid filter string")
				}
				pFilter.MatchInstance = *m
			}
		} else {
			pFilter.MatchInstance = *regexp.MustCompile(".*")
		}

		if len(s) >= 3 {
			switch action := strings.ToLower(s[2]); action {
			case "sum", "add":
				pFilter.Action = PerfFilterActionSum
			case "drop":
				pFilter.Action = PerfFilterActionDrop
			case "split":
				pFilter.Action = PerfFilterActionSpit
			case "rename":
				pFilter.Action = PerfFilterActionRename
			case "instance-rename":
				pFilter.Action = PerfFilterActionInstanceRename
			default:
				return nil, fmt.Errorf("invalid filter string")
			}
		} else {
			pFilter.Action = PerfFilterActionDrop
		}

		if len(s) >= 4 {
			pFilter.NewName = s[3]
		}

		pFilters = append(pFilters, pFilter)
	}
	return pFilters, nil
}

func (c PerfSensorConfig) MustParseFilters() []PerfFilter {
	filters, err := c.ParseFilters()
	if err != nil {
		panic(err)
	}
	return filters
}

func (c *PerfSensorConfig) Validate() error {
	_, err := c.ParseFilters()
	return err
}

type BackendConfig struct {
	Type  string
	Redis RedisConfig
}

type RedisConfig struct {
	Address  string
	Password string
	Index    int
}

func (c *PerfSensorConfig) SensorConfig() SensorConfig {
	return SensorConfig{
		Enabled:         c.Enabled,
		MaxAge:          c.MaxAge,
		RefreshInterval: c.RefreshInterval,
		RefreshTimeout:  c.RefreshTimeout,
	}
}

type TagsSensorConfig struct {
	SensorConfig
	CategoryToCollect []string
}

func DefaultScraperConfig() ScraperConfig {
	return ScraperConfig{
		Cluster: SensorConfig{
			Enabled:         true,
			MaxAge:          120 * time.Second,
			RefreshInterval: 60 * time.Second,
		},
		ComputeResource: SensorConfig{
			Enabled:         true,
			MaxAge:          120 * time.Second,
			RefreshInterval: 60 * time.Second,
		},
		Datastore: SensorConfig{
			Enabled:         true,
			MaxAge:          120 * time.Second,
			RefreshInterval: 30 * time.Second,
		},
		Host: SensorConfig{
			Enabled:         true,
			MaxAge:          120 * time.Second,
			RefreshInterval: 30 * time.Second,
		},
		ResourcePool: SensorConfig{
			Enabled:         true,
			MaxAge:          120 * time.Second,
			RefreshInterval: 60 * time.Second,
		},
		Datacenter: SensorConfig{
			Enabled:         true,
			MaxAge:          120 * time.Second,
			RefreshInterval: 60 * time.Second,
		},
		Folder: SensorConfig{
			Enabled:         true,
			MaxAge:          120 * time.Second,
			RefreshInterval: 60 * time.Second,
		},
		Spod: SensorConfig{
			Enabled:         true,
			MaxAge:          120 * time.Second,
			RefreshInterval: 60 * time.Second,
		},
		Tags: TagsSensorConfig{
			SensorConfig: SensorConfig{
				Enabled:         true,
				MaxAge:          600 * time.Second,
				RefreshInterval: 290 * time.Second,
			},
			CategoryToCollect: []string{},
		},
		VirtualMachine: SensorConfig{
			Enabled:         true,
			MaxAge:          120 * time.Second,
			RefreshInterval: 60 * time.Second,
		},
		HostPerf: PerfSensorConfig{
			Enabled:         true,
			MaxAge:          10 * time.Minute,
			RefreshInterval: 60 * time.Second,
			MaxSampleWindow: 5 * time.Minute,
			SampleInterval:  20 * time.Second,
			DefaultMetrics:  true,
		},
		VirtualMachinePerf: PerfSensorConfig{
			Enabled:         true,
			MaxAge:          10 * time.Minute,
			RefreshInterval: 60 * time.Second,
			MaxSampleWindow: 5 * time.Minute,
			SampleInterval:  20 * time.Second,
			DefaultMetrics:  true,
		},
		Backend: BackendConfig{
			Type: "memory",
			Redis: RedisConfig{
				Address:  "localhost:6379",
				Password: "",
				Index:    0,
			},
		},

		ClientPoolSize: 5,
	}
}

func (c ScraperConfig) Validate() error {
	if !regexp.MustCompile(URL_REGEX_PATTERN).MatchString(c.VCenter) {
		return fmt.Errorf("invalid URL %s", c.VCenter)
	}

	if c.ClientPoolSize <= 0 {
		return fmt.Errorf("ClientPoolSize cannot be smaller than 1")
	}

	if !(c.Backend.Type == "memory" || c.Backend.Type == "redis") {
		return fmt.Errorf("backend.type must be one of [memory|redis]")
	}

	if !c.Host.Enabled && c.VirtualMachine.Enabled {
		return fmt.Errorf(`HostScraperEnabled must be enabled when 
VirtualMachineScraperEnabled is enabled because scraper needs the hosts 
when it queries the vm's`)
	}

	if !c.Datacenter.Enabled && c.Host.Enabled {
		return fmt.Errorf(`DatacenterSensor must be enabled when 
HostSensor is enabled because scraper needs the dc's 
when it queries the hosts`)
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

	if err := c.HostPerf.Validate(); err != nil {
		return fmt.Errorf("invalid hostperf config: %v", err)
	}

	if err := c.VirtualMachinePerf.Validate(); err != nil {
		return fmt.Errorf("invalid vmperf config: %v", err)
	}
	return nil
}

func (c ScraperConfig) Endpoint() string {
	re := regexp.MustCompile(URL_REGEX_PATTERN)
	groups := re.FindStringSubmatch(c.VCenter)
	if len(groups) != 5 {
		panic("Invalid URL")
	}
	scheme := groups[1]
	if scheme == "" {
		scheme = "https"
	}
	host := groups[2]
	var port int
	if groups[3] != "" {
		port, _ = strconv.Atoi(groups[3])
	} else {
		switch scheme {
		case "https":
			port = 443
		case "http":
			port = 80
		}
	}
	path := groups[4]
	if path == "/" {
		path = ""
	} else if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	return fmt.Sprintf("%s://%s:%d%s", scheme, host, port, path)
}

func (c ScraperConfig) URL() (*url.URL, error) {
	return url.Parse(c.Endpoint())
}

func (c ScraperConfig) SoapURL() (*url.URL, error) {
	u, err := soap.ParseURL(c.Endpoint())
	if err != nil {
		return nil, fmt.Errorf("unable to parse url %s", c.Endpoint())
	}
	return u, err
}
