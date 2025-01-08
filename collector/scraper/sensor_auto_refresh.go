package scraper

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/intrinsec/govc_exporter/collector/pool"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

var SensorSkipRefreshError = errors.New("refresh is still running, skipping this one")

type VMwareResource interface {
	mo.VirtualMachine | mo.ComputeResource | mo.HostSystem | mo.StoragePod | mo.Datastore | mo.ManagedEntity | mo.ResourcePool
}

type cacheConfig struct {
	MaxAge             int
	RefreshInterval    int
	CleanCacheInterval int
}

type AutoRefreshSensor[T VMwareResource] struct {
	objects       map[types.ManagedObjectReference]VMwareCacheItem[T]
	refreshFunc   func(*govmomi.Client, *slog.Logger) ([]T, error)
	lock          sync.Mutex
	refreshLock   sync.Mutex
	config        cacheConfig
	metrics       *SensorMetrics
	refreshTicker *time.Ticker
	cleanupTicker *time.Ticker
}

func NewAutoRefreshSensor[T VMwareResource](refreshFunc func(*govmomi.Client, *slog.Logger) ([]T, error), conf cacheConfig) *AutoRefreshSensor[T] {
	return &AutoRefreshSensor[T]{
		objects:     map[types.ManagedObjectReference]VMwareCacheItem[T]{},
		refreshFunc: refreshFunc,
		config:      conf,
	}
}

func (o *AutoRefreshSensor[T]) GetMetrics() SensorMetrics {
	return *o.metrics
}

func (o *AutoRefreshSensor[T]) getTypeString() string {
	split := strings.Split(fmt.Sprint(reflect.TypeOf(o)), ".")
	return strings.TrimRight(split[len(split)-1], "]")
}

func (o *AutoRefreshSensor[T]) Add(kind types.ManagedObjectReference, r *T) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.objects[kind] = *NewVMwareCacheItem(r, time.Duration(o.config.MaxAge)*time.Second)
}

func (o *AutoRefreshSensor[T]) GetAll() []*T {
	result := []*T{}
	o.lock.Lock()
	defer o.lock.Unlock()
	for _, v := range o.objects {
		result = append(result, v.Item)
	}
	return result
}

func (o *AutoRefreshSensor[T]) GetAllRefs() []types.ManagedObjectReference {
	result := []types.ManagedObjectReference{}
	o.lock.Lock()
	defer o.lock.Unlock()
	if len(o.objects) == 0 {
		return nil
	}

	for k := range o.objects {
		result = append(result, k)
	}
	return result
}

func (o *AutoRefreshSensor[t]) Get(ref types.ManagedObjectReference) *t {
	o.lock.Lock()
	defer o.lock.Unlock()
	if o, ok := o.objects[ref]; ok {
		return o.Item
	}
	return nil
}

func (o *AutoRefreshSensor[t]) Refresh(clientPool pool.Pool[govmomi.Client], logger *slog.Logger) error {
	t1 := time.Now()
	client, clientId := clientPool.Acquire()
	defer clientPool.Release(clientId)

	t2 := time.Now()
	if !o.refreshLock.TryLock() {
		return SensorSkipRefreshError
	}
	defer o.refreshLock.Unlock()
	results, err := o.refreshFunc(client, logger)
	t3 := time.Now()
	if err != nil {
		o.metrics = &SensorMetrics{
			Name:           o.getTypeString(),
			QueryTime:      t2.Sub(t1),
			ClientWaitTime: 0,
			Status:         false,
		}
		return err
	}
	for _, r := range results {
		switch p := any(r).(type) {
		case mo.HostSystem:
			o.Add(p.ManagedEntity.Self, &r)
		case mo.ClusterComputeResource:
			o.Add(p.ComputeResource.ManagedEntity.Self, &r)
		case mo.ComputeResource:
			o.Add(p.ManagedEntity.Self, &r)
		case mo.VirtualMachine:
			o.Add(p.ManagedEntity.Self, &r)
		case mo.StoragePod:
			o.Add(p.ManagedEntity.Self, &r)
		case mo.Datastore:
			o.Add(p.ManagedEntity.Self, &r)
		case mo.ResourcePool:
			o.Add(p.ManagedEntity.Self, &r)
		default:
			panic(fmt.Sprintf("invalid type: %s", reflect.TypeOf(p)))
		}
	}

	o.metrics = &SensorMetrics{
		QueryTime:      t2.Sub(t1),
		ClientWaitTime: t3.Sub(t2),
		Status:         true,
	}
	return nil
}

func (o *AutoRefreshSensor[T]) clean() {
	o.lock.Lock()
	defer o.lock.Unlock()
	newMap := map[types.ManagedObjectReference]VMwareCacheItem[T]{}
	for k, v := range o.objects {
		if !v.Expired() {
			newMap[k] = v
		}
	}
	o.objects = newMap
}

func (o *AutoRefreshSensor[T]) Start(clientPool pool.Pool[govmomi.Client], logger *slog.Logger) {
	o.refreshTicker = time.NewTicker(time.Duration(o.config.RefreshInterval) * time.Second)
	o.cleanupTicker = time.NewTicker(time.Duration(o.config.CleanCacheInterval) * time.Second)

	go func() {
		for ; true; <-o.refreshTicker.C {
			err := o.Refresh(clientPool, logger)
			if err == nil {
				logger.Info("refresh successfull", "sensor_type", o.getTypeString())
			} else if errors.Is(err, SensorSkipRefreshError) {
				logger.Warn("Skipping sensor refresh", "err", err.Error(), "sensor_type", o.getTypeString())
			} else {
				logger.Warn("Failed to refresh sensor", "err", err.Error(), "sensor_type", o.getTypeString())
			}
		}
	}()

	go func() {
		for range o.cleanupTicker.C {
			o.clean()
			logger.Debug("clean successfull", "sensor_type", o.getTypeString())
		}
	}()
}

func (o *AutoRefreshSensor[T]) Stop(logger *slog.Logger) {
	logger.Info("stopping refresh ticker...", "sensor_type", o.getTypeString())
	o.refreshTicker.Stop()
	logger.Info("refresh ticker stopped", "sensor_type", o.getTypeString())

	logger.Info("stopping cleanup ticker...", "sensor_type", o.getTypeString())
	o.cleanupTicker.Stop()
	logger.Info("cleanup ticker stopped", "sensor_type", o.getTypeString())
}
