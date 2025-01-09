package scraper

import (
	"log/slog"
	"sync"
	"time"

	"github.com/intrinsec/govc_exporter/collector/pool"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type OnDemandSensor struct {
	objects       map[types.ManagedObjectReference]VMwareCacheItem[mo.ManagedEntity]
	lock          sync.Mutex
	getFunc       func(types.ManagedObjectReference, pool.Pool[govmomi.Client], *slog.Logger) (*mo.ManagedEntity, error)
	config        sensorConfig
	logger        *slog.Logger
	cleanupTicker *time.Ticker
	clientPool    *pool.Pool[govmomi.Client]
}

func NewOnDemandSensor(getFunc func(types.ManagedObjectReference, pool.Pool[govmomi.Client], *slog.Logger) (*mo.ManagedEntity, error), conf sensorConfig) *OnDemandSensor {
	return &OnDemandSensor{
		objects: map[types.ManagedObjectReference]VMwareCacheItem[mo.ManagedEntity]{},
		getFunc: getFunc,
		config:  conf,
	}
}

func (o *OnDemandSensor) Add(kind types.ManagedObjectReference, r *mo.ManagedEntity) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.objects[kind] = *NewVMwareCacheItem(r, time.Duration(o.config.MaxAge)*time.Second)
}

func (o *OnDemandSensor) getTypeString() string {
	return "OnDemand"
}

func (o *OnDemandSensor) GetAll() []*mo.ManagedEntity {
	result := []*mo.ManagedEntity{}
	o.lock.Lock()
	defer o.lock.Unlock()
	for _, v := range o.objects {
		result = append(result, v.Item)
	}
	return result
}

func (o *OnDemandSensor) GetAllRefs() []types.ManagedObjectReference {
	result := []types.ManagedObjectReference{}
	o.lock.Lock()
	defer o.lock.Unlock()
	for k := range o.objects {
		result = append(result, k)
	}
	return result
}

func (o *OnDemandSensor) Get(ref types.ManagedObjectReference) *mo.ManagedEntity {
	entity, hasInCache := func() (*mo.ManagedEntity, bool) {
		o.lock.Lock()
		defer o.lock.Unlock()
		if entiry, ok := o.objects[ref]; ok {
			return entiry.Item, true
		} else {
			return nil, false
		}
	}()

	if hasInCache {
		return entity
	}

	entity, err := o.getFunc(ref, *o.clientPool, o.logger)
	if err != nil {
		return nil
	}

	o.Add(ref, entity)

	return entity
}

func (o *OnDemandSensor) clean() {
	o.lock.Lock()
	defer o.lock.Unlock()
	newMap := map[types.ManagedObjectReference]VMwareCacheItem[mo.ManagedEntity]{}
	for k, v := range o.objects {
		if !v.Expired() {
			newMap[k] = v
		}
	}
	o.objects = newMap
}

func (o *OnDemandSensor) Start(clientPool pool.Pool[govmomi.Client], logger *slog.Logger) {
	o.logger = logger
	o.clientPool = &clientPool

	o.cleanupTicker = time.NewTicker(time.Duration(o.config.CleanCacheInterval) * time.Second)

	go func() {
		for range o.cleanupTicker.C {
			o.clean()
			logger.Debug("clean successfull", "sensor_type", o.getTypeString())
		}
	}()
}

func (o *OnDemandSensor) Stop(logger *slog.Logger) {
	logger.Info("stopping cleanup ticker...", "sensor_type", o.getTypeString())
	o.cleanupTicker.Stop()
	logger.Info("cleanup ticker stopped", "sensor_type", o.getTypeString())
}
