package collector

import (
	"context"
	"sync"
	"time"

	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type VMwareResource interface {
	mo.VirtualMachine | mo.ClusterComputeResource | mo.HostSystem
}

type CacheObject[T VMwareResource] struct {
	objects            map[types.ManagedObjectReference]VMwareCacheItem[T]
	refreshFunc        func() ([]T, error)
	lock               sync.Mutex
	MaxAgeSec          int
	RefreshIntervalSec int
}

func NewCacheObject[T VMwareResource](refreshFunc func() ([]T, error), maxAgeSec int, refreshIntervalSec int) *CacheObject[T] {
	return &CacheObject[T]{
		objects:            map[types.ManagedObjectReference]VMwareCacheItem[T]{},
		refreshFunc:        refreshFunc,
		MaxAgeSec:          maxAgeSec,
		RefreshIntervalSec: refreshIntervalSec,
	}
}

func (o *CacheObject[T]) Add(kind types.ManagedObjectReference, r *T) {
	o.lock.Lock()
	defer o.lock.Unlock()
	maxAge := time.Duration(o.MaxAgeSec) * time.Second
	o.objects[kind] = *NewVMwareCacheItem(r, maxAge)
}

func (o *CacheObject[T]) GetAll() []*T {
	result := []*T{}
	o.lock.Lock()
	defer o.lock.Unlock()
	for _, v := range o.objects {
		result = append(result, v.Item)
	}
	return result
}

func (o *CacheObject[t]) Get(ref types.ManagedObjectReference) *t {
	o.lock.Lock()
	defer o.lock.Unlock()
	if o, ok := o.objects[ref]; ok {
		return o.Item
	}
	return nil
}

func (o *CacheObject[t]) Refresh() error {
	results, err := o.refreshFunc()
	if err != nil {
		return err
	}
	for _, r := range results {
		switch p := any(r).(type) {
		case mo.HostSystem:
			o.Add(p.ManagedEntity.Self, &r)
		case mo.ClusterComputeResource:
			o.Add(p.ManagedEntity.Self, &r)
		case mo.VirtualMachine:
			o.Add(p.ManagedEntity.Self, &r)
		}
	}
	return nil
}

func (o *CacheObject[T]) clean() {
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

func (o *CacheObject[T]) Start(ctx context.Context, errChan chan error) {
	refreshTicker := time.NewTicker(time.Duration(o.RefreshIntervalSec))
	go func() {
		for range refreshTicker.C {
			errChan <- o.Refresh()
		}
	}()
	cleanupTicker := time.NewTicker(time.Duration(10) * time.Second)
	go func() {
		for range cleanupTicker.C {
			o.clean()
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				refreshTicker.Stop()
				cleanupTicker.Stop()
				return
			default:
			}
		}
	}()
}
