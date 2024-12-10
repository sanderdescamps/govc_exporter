package collector

import (
	"time"
)

const (
	DEFAULT_HOST_MAX_AGE_SEC    = 5 * 60
	DEFAULT_CLUSTER_MAX_AGE_SEC = 10 * 60
	DEFAULT_VM_MAX_AGE_SEC      = 5 * 60

	DEFAULT_CLEAN_CACHE_INTERVAL_SEC = 10

	DEFAULT_HOST_REFRESH_INTERVAL_SEC    = 30
	DEFAULT_CLUSTER_REFRESH_INTERVAL_SEC = 30
	DEFAULT_VM_REFRESH_INTERVAL_SEC      = 5 * 60
)

type VMwareConfig struct {
	ClientPoolSize            int    `json:"client_pool_size"`
	Endpoint                  string `json:"endpoint"`
	Username                  string `json:"username"`
	Password                  string `json:"password"`
	HostMaxAgeSec             int    `json:"host_max_age_sec"`
	ClusterMaxAgeSec          int    `json:"cluster_max_age_sec"`
	VmMaxAgeSec               int    `json:"vm_max_age_sec"`
	HostRefreshIntervalSec    int    `json:"host_refresh_interval_sec"`
	ClusterRefreshIntervalSec int    `json:"cluster_refresh_interval_sec"`
	VmRefreshIntervalSec      int    `json:"vm_refresh_interval_sec"`
	CleanCacheIntervalSec     int    `json:"clean_cache_interval_sec"`
}

func (c VMwareConfig) HostMaxAge() time.Duration {
	maxAge := c.HostMaxAgeSec
	if maxAge == 0 {
		maxAge = DEFAULT_MAX_AGE_SEC
	}
	return time.Duration(maxAge) * time.Second
}

func (c VMwareConfig) ClusterMaxAge() time.Duration {
	maxAge := c.ClusterMaxAgeSec
	if maxAge == 0 {
		maxAge = DEFAULT_MAX_AGE_SEC
	}
	return time.Duration(maxAge) * time.Second
}

func (c VMwareConfig) VmMaxAge() time.Duration {
	maxAge := c.VmMaxAgeSec
	if maxAge == 0 {
		maxAge = DEFAULT_MAX_AGE_SEC
	}
	return time.Duration(maxAge) * time.Second
}

func (c VMwareConfig) HostRefreshInterval() time.Duration {
	maxAge := c.HostRefreshIntervalSec
	if maxAge == 0 {
		maxAge = DEFAULT_HOST_REFRESH_INTERVAL_SEC
	}
	return time.Duration(maxAge) * time.Second
}

func (c VMwareConfig) ClusterRefreshInterval() time.Duration {
	maxAge := c.ClusterRefreshIntervalSec
	if maxAge == 0 {
		maxAge = DEFAULT_CLUSTER_REFRESH_INTERVAL_SEC
	}
	return time.Duration(maxAge) * time.Second
}

func (c VMwareConfig) VmRefreshInterval() time.Duration {
	maxAge := c.VmRefreshIntervalSec
	if maxAge == 0 {
		maxAge = DEFAULT_VM_REFRESH_INTERVAL_SEC
	}
	return time.Duration(maxAge) * time.Second
}

func (c VMwareConfig) CleanCacheInterval() time.Duration {
	maxAge := c.VmRefreshIntervalSec
	if maxAge == 0 {
		maxAge = DEFAULT_CLEAN_CACHE_INTERVAL_SEC
	}
	return time.Duration(maxAge) * time.Second
}
