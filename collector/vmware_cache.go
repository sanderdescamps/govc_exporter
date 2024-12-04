package collector

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type VMwareConfig struct {
	RefreshPeriod  int    `json:"refresh_interval"`
	ClientPoolSize int    `json:"client_pool_size"`
	Endpoint       string `json:"endpoint"`
	Username       string `json:"username"`
	Password       string `json:"password"`
}

type VMwareActiveCache struct {
	cache              map[types.ManagedObjectReference]VMwareCacheItem
	sensors            []func() ([]mo.ManagedEntity, error)
	config             VMwareConfig
	clientPool         Pool
	cacheLock          sync.Mutex
	refreshLock        sync.Mutex
	cacheRefreshTicker *time.Ticker
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

func (c *VMwareActiveCache) GetAllWithKindFromCache(kind string, dst interface{}) error {
	results := []mo.ManagedEntity{}

	for _, i := range c.cache {
		if i.Entity.Self.Type == kind {
			results = append(results, i.Entity)
		}
	}

	c1 := reflect.ValueOf(results)
	reflect.ValueOf(dst).Elem().Set(c1)

	return nil
}

func (c *VMwareActiveCache) getFromCache(obj types.ManagedObjectReference, dst interface{}) (bool, error) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()
	if e, ok := c.cache[obj]; ok {
		c1 := reflect.ValueOf(e)
		reflect.ValueOf(dst).Elem().Set(c1)
		return true, nil
	}

	return false, nil
}

func (c *VMwareActiveCache) Get(ctx context.Context, obj types.ManagedObjectReference, ps []string, dst interface{}) error {
	if e, err := c.getFromCache(obj, dst); e {
		return nil
	} else if err != nil {
		return err
	}

	client := c.clientPool.GetClient()
	pc := property.DefaultCollector(client.Client)

	var temp mo.ManagedEntity
	pc.RetrieveOne(ctx, obj, ps, temp)

	c1 := reflect.ValueOf(temp)
	reflect.ValueOf(dst).Elem().Set(c1)

	return nil
}

func (c *VMwareActiveCache) cleanCache() {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()
	clean_cache := map[types.ManagedObjectReference]VMwareCacheItem{}
	for k, v := range c.cache {
		if !v.Expired(5 * time.Minute) {
			clean_cache[k] = v
		}
	}
	c.cache = clean_cache
}

func (c *VMwareActiveCache) refreshSensors() error {
	for _, s := range c.sensors {
		entities, err := s()
		if err != nil {
			return err
		}

		func() {
			c.cacheLock.Lock()
			defer c.cacheLock.Unlock()
			for _, e := range entities {
				c.cache[e.Self] = VMwareCacheItem{
					Entity:    e,
					timeStamp: time.Now(),
				}
			}
		}()
	}
	return nil
}

func (c *VMwareActiveCache) Start() {
	c.cacheRefreshTicker = time.NewTicker(time.Duration(c.config.RefreshPeriod) * time.Second)
	go func() {
		for range c.cacheRefreshTicker.C {
			c.refreshSensors()
			c.cleanCache()
		}
	}()
}

func (c *VMwareActiveCache) Stop() {
	c.cacheRefreshTicker.Stop()
}
