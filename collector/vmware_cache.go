package collector

import (
	"context"
	"sync"
	"time"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type VMwareActiveCache struct {
	cache                 map[types.ManagedObjectReference]interface{}
	sensors               map[string]func() ([]mo.ManagedEntity, error)
	config                VMwareConfig
	clientPool            Pool
	cacheLock             sync.Mutex
	sensorMaintenanceLock sync.Mutex
	refreshLock           sync.Mutex
	cacheRefreshTicker    *time.Ticker

	hostRefreshTicker    *time.Ticker
	hostCacheLock        sync.Mutex
	hostCache            map[types.ManagedObjectReference]*VMwareCacheItem[mo.HostSystem]
	clusterRefreshTicker *time.Ticker
	clusterCacheLock     sync.Mutex
	clusterCache         map[types.ManagedObjectReference]*VMwareCacheItem[mo.ClusterComputeResource]
	vmRefreshTicker      *time.Ticker
	vmCacheLock          sync.Mutex
	vmCache              map[types.ManagedObjectReference]*VMwareCacheItem[mo.VirtualMachine]
}

func NewVMwareActiveCache(conf VMwareConfig) *VMwareActiveCache {
	ctx := context.Background()
	clientConf := ClientConf{
		Endpoint: conf.Endpoint,
		Username: conf.Username,
		Password: conf.Password,
	}
	pool := NewVMwareClientPool(ctx, conf.ClientPoolSize, clientConf)

	return &VMwareActiveCache{
		clientPool: pool,
	}
}

func (c *VMwareActiveCache) refreshHosts() error {
	var items []mo.HostSystem

	client := c.clientPool.GetClient()
	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"HostSystem"},
		true,
	)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"HostSystem"},
		[]string{
			"parent",
			"summary",
		},
		&items,
	)

	c.hostCacheLock.Lock()
	defer c.hostCacheLock.Unlock()
	for _, i := range items {
		c.hostCache[i.ManagedEntity.Self] = NewVMwareCacheItem(i, c.config.HostMaxAge())
	}
	return nil
}

func (c *VMwareActiveCache) refreshClusters() error {
	var items []mo.ClusterComputeResource

	client := c.clientPool.GetClient()
	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"ClusterComputeResource"},
		true,
	)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{"ClusterComputeResource"},
		[]string{
			// "parent",
			// "summary",
		},
		&items,
	)

	c.clusterCacheLock.Lock()
	defer c.clusterCacheLock.Unlock()
	for _, i := range items {
		c.clusterCache[i.ManagedEntity.Self] = NewVMwareCacheItem(i, c.config.ClusterMaxAge())
	}
	return nil
}

func (c *VMwareActiveCache) refreshVMbyHost(host types.ManagedObjectReference) error {
	var items []mo.VirtualMachine
	client := c.clientPool.GetClient()
	ctx := context.Background()
	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{"VirtualMachine"},
		true,
	)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	err = v.RetrieveWithFilter(
		ctx,
		[]string{"VirtualMachine"},
		[]string{
			"config",
			//"datatore",
			"guest",
			"guestHeartbeatStatus",
			"network",
			"parent",
			"resourceConfig",
			"resourcePool",
			"runtime",
			"snapshot",
			"summary",
		},
		&items, property.Filter{"Runtime.Host": host},
	)
	c.vmCacheLock.Lock()
	defer c.vmCacheLock.Unlock()
	for _, i := range items {
		c.vmCache[i.ManagedEntity.Self] = NewVMwareCacheItem(i, c.config.VmMaxAge())
	}
	return nil
}

func (c *VMwareActiveCache) refreshVMs() error {
	hostRefs := c.getHostRefs()

	wg := sync.WaitGroup{}
	wg.Add(len(hostRefs))
	errChan := make(chan error, 1)
	for _, h := range hostRefs {
		go func() {
			defer wg.Done()
			errChan <- c.refreshVMbyHost(h)

		}()
	}
	wg.Wait()
	return <-errChan
}

func (c *VMwareActiveCache) getHostRefs() []types.ManagedObjectReference {
	hostRefs := []types.ManagedObjectReference{}
	c.hostCacheLock.Lock()
	defer c.hostCacheLock.Unlock()
	for k, _ := range c.hostCache {
		hostRefs = append(hostRefs, k)
	}
	return hostRefs
}

func (c *VMwareActiveCache) GetVMs() []mo.VirtualMachine {
	result := []mo.VirtualMachine{}
	c.vmCacheLock.Lock()
	defer c.vmCacheLock.Unlock()
	for _, cacheItem := range c.vmCache {
		if !cacheItem.Expired() {
			result = append(result, cacheItem.Item)
		}
	}
	return result
}

func (c *VMwareActiveCache) GetClusters() []mo.ClusterComputeResource {
	result := []mo.ClusterComputeResource{}
	c.clusterCacheLock.Lock()
	defer c.clusterCacheLock.Unlock()
	for _, cacheItem := range c.clusterCache {
		if !cacheItem.Expired() {
			result = append(result, cacheItem.Item)
		}
	}
	return result
}

func (c *VMwareActiveCache) GetHosts() []mo.HostSystem {
	result := []mo.HostSystem{}
	c.hostCacheLock.Lock()
	defer c.hostCacheLock.Unlock()
	for _, cacheItem := range c.hostCache {
		if !cacheItem.Expired() {
			result = append(result, cacheItem.Item)
		}
	}
	return result
}

// func (c *VMwareActiveCache) GetAllWithKindFromCache(kind string, dst interface{}) error {
// 	results := []mo.ManagedEntity{}

// 	for _, i := range c.cache {
// 		if i.Entity.Self.Type == kind {
// 			results = append(results, i.Entity)
// 		}
// 	}

// 	c1 := reflect.ValueOf(results)
// 	reflect.ValueOf(dst).Elem().Set(c1)

// 	return nil
// }

// func (c *VMwareActiveCache) getFromCache(obj types.ManagedObjectReference, dst interface{}) (bool, error) {
// 	c.cacheLock.Lock()
// 	defer c.cacheLock.Unlock()
// 	if e, ok := c.cache[obj]; ok {
// 		c1 := reflect.ValueOf(e)
// 		reflect.ValueOf(dst).Elem().Set(c1)
// 		return true, nil
// 	}

// 	return false, nil
// }

// func (c *VMwareActiveCache) Get(ctx context.Context, obj types.ManagedObjectReference, ps []string, dst interface{}) error {
// 	if e, err := c.getFromCache(obj, dst); e {
// 		return nil
// 	} else if err != nil {
// 		return err
// 	}

// 	client := c.clientPool.GetClient()
// 	pc := property.DefaultCollector(client.Client)

// 	var temp mo.ManagedEntity
// 	pc.RetrieveOne(ctx, obj, ps, temp)

// 	c1 := reflect.ValueOf(temp)
// 	reflect.ValueOf(dst).Elem().Set(c1)

// 	return nil
// }

func (c *VMwareActiveCache) cleanCache() {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()
	clean_cache := map[types.ManagedObjectReference]VMwareCacheItem{}
	for k, v := range c.cache {
		if !v.Expired() {
			clean_cache[k] = v
		}
	}
	c.cache = clean_cache
}

// func (c *VMwareActiveCache) refreshSensors() error {
// 	for _, s := range c.sensors {
// 		entities, err := s()
// 		if err != nil {
// 			return err
// 		}

// 		func() {
// 			c.cacheLock.Lock()
// 			defer c.cacheLock.Unlock()
// 			for _, e := range entities {
// 				c.cache[e.Self] = VMwareCacheItem{
// 					Entity:    e,
// 					timeStamp: time.Now(),
// 				}
// 			}
// 		}()
// 	}
// 	return nil
// }

func (c *VMwareActiveCache) Start(ctx context.Context, errChan chan error) {
	hostRefreshTicker := time.NewTicker(c.config.HostRefreshInterval())
	go func() {
		for range hostRefreshTicker.C {
			errChan <- c.refreshHosts()
		}
	}()

	clusterRefreshTicker := time.NewTicker(c.config.HostRefreshInterval())
	go func() {
		for range clusterRefreshTicker.C {
			errChan <- c.refreshHosts()
		}
	}()

	vmRefreshTicker := time.NewTicker(c.config.HostRefreshInterval())
	go func() {
		for range vmRefreshTicker.C {
			errChan <- c.refreshHosts()
		}
	}()

	cleanupTicker := time.NewTicker(c.config.CleanCacheInterval())
	go func() {
		for range cleanupTicker.C {
			errChan <- c.cleanCache()
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.hostRefreshTicker.Stop()
				c.clusterRefreshTicker.Stop()
				c.vmRefreshTicker.Stop()
				cleanupTicker.Stop()
				return
			default:
			}
			// do work
		}
	}()
}

func (c *VMwareActiveCache) Stop() {
	c.cacheRefreshTicker.Stop()
}
