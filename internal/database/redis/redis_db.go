package redis_db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"reflect"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
)

type DB struct {
	client   *redis.Client
	password string
	addr     string
	dbIndex  int
}

func NewDB(addr string, password string, index int) *DB {
	db := &DB{
		addr:     addr,
		password: password,
		dbIndex:  index,
	}
	return db
}

func (db *DB) Connect(ctx context.Context) error {
	if db.client != nil {
		return nil
	}

	db.client = redis.NewClient(&redis.Options{
		Addr:     db.addr,
		Password: db.password, // no password set
		DB:       db.dbIndex,  // use default DB
	})

	return nil
}

func (db *DB) Disconnect(ctx context.Context) error {
	db.client = nil
	return nil
}

func (db *DB) Set(ctx context.Context, typ objects.ManagedObjectTypes, key string, obj interface{}, ttl time.Duration) error {
	db.Connect(ctx)

	redisKey := fmt.Sprintf("%s:%s", typ.String(), key)
	status := db.client.JSONSet(ctx, redisKey, "$", obj)
	if status.Err() != nil {
		return status.Err()
	}

	if ttl != 0 {
		_, err := db.client.Expire(ctx, redisKey, ttl).Result()
		if err != nil {
			return fmt.Errorf("failed to set expiration of object: %v", err)
		}
	}
	return nil
}

func (db *DB) SetObj(ctx context.Context, ref objects.ManagedObjectReference, obj interface{}, ttl time.Duration) error {
	db.Connect(ctx)

	return db.Set(ctx, ref.Type, ref.ID(), obj, ttl)
}

func (db *DB) GetObj(ctx context.Context, ref objects.ManagedObjectReference, res interface{}) error {
	return db.Get(ctx, ref.Type, ref.ID(), res)
}

func (db *DB) Get(ctx context.Context, typ objects.ManagedObjectTypes, key string, res interface{}) error {
	resv := reflect.ValueOf(res)
	if resv.Kind() != reflect.Pointer || resv.IsNil() {
		return errors.New("input error. Value must be a pointer")
	}

	var redisKey string = func() string {
		if !strings.HasPrefix(key, typ.String()+":") {
			return fmt.Sprintf("%s:%s", typ.String(), key)
		}
		return key
	}()

	redisCmd := db.client.JSONGet(ctx, redisKey, ".")
	if redisCmd.Err() != nil || redisCmd.Val() == "" {
		return nil
	}

	if err := json.Unmarshal([]byte(redisCmd.Val()), res); err != nil {
		return err
	}

	return nil
}

func (db *DB) SetCluster(ctx context.Context, cluster objects.Cluster, ttl time.Duration) error {
	err := db.SetObj(ctx, cluster.Self, cluster, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetComputeResource(ctx context.Context, compResource objects.ComputeResource, ttl time.Duration) error {
	err := db.SetObj(ctx, compResource.Self, compResource, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetDatacenter(ctx context.Context, dc objects.Datacenter, ttl time.Duration) error {
	err := db.SetObj(ctx, dc.Self, dc, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetDatastore(ctx context.Context, ds objects.Datastore, ttl time.Duration) error {
	err := db.SetObj(ctx, ds.Self, ds, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetFolder(ctx context.Context, f objects.Folder, ttl time.Duration) error {
	err := db.SetObj(ctx, f.Self, f, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetHost(ctx context.Context, host objects.Host, ttl time.Duration) error {
	err := db.SetObj(ctx, host.Self, host, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetStoragePod(ctx context.Context, spod objects.StoragePod, ttl time.Duration) error {
	err := db.SetObj(ctx, spod.Self, spod, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetResourcePool(ctx context.Context, rp objects.ResourcePool, ttl time.Duration) error {
	err := db.SetObj(ctx, rp.Self, rp, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetVM(ctx context.Context, vm objects.VirtualMachine, ttl time.Duration) error {
	err := db.SetObj(ctx, vm.Self, vm, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetCluster(ctx context.Context, ref objects.ManagedObjectReference) *objects.Cluster {
	var cluster objects.Cluster
	err := db.GetObj(ctx, ref, &cluster)
	if err != nil {
		return nil
	}
	return &cluster
}

func (db *DB) GetComputeResource(ctx context.Context, ref objects.ManagedObjectReference) *objects.ComputeResource {
	var compResource objects.ComputeResource
	err := db.GetObj(ctx, ref, &compResource)
	if err != nil {
		return nil
	}
	return &compResource
}

func (db *DB) GetDatacenter(ctx context.Context, ref objects.ManagedObjectReference) *objects.Datacenter {
	var dc objects.Datacenter
	err := db.GetObj(ctx, ref, &dc)
	if err != nil {
		return nil
	}
	return &dc
}

func (db *DB) GetDatastore(ctx context.Context, ref objects.ManagedObjectReference) *objects.Datastore {
	var ds objects.Datastore
	err := db.GetObj(ctx, ref, &ds)
	if err != nil {
		return nil
	}
	return &ds
}

func (db *DB) GetFolder(ctx context.Context, ref objects.ManagedObjectReference) *objects.Folder {
	var folder objects.Folder
	err := db.GetObj(ctx, ref, &folder)
	if err != nil {
		return nil
	}
	return &folder
}

func (db *DB) GetHost(ctx context.Context, ref objects.ManagedObjectReference) *objects.Host {
	var host objects.Host
	err := db.GetObj(ctx, ref, &host)
	if err != nil {
		return nil
	}
	return &host
}

func (db *DB) GetStoragePod(ctx context.Context, ref objects.ManagedObjectReference) *objects.StoragePod {
	var spod objects.StoragePod
	err := db.GetObj(ctx, ref, &spod)
	if err != nil {
		return nil
	}
	return &spod
}

func (db *DB) GetResourcePool(ctx context.Context, ref objects.ManagedObjectReference) *objects.ResourcePool {
	var rp objects.ResourcePool
	err := db.GetObj(ctx, ref, &rp)
	if err != nil {
		return nil
	}
	return &rp
}

func (db *DB) GetVM(ctx context.Context, ref objects.ManagedObjectReference) *objects.VirtualMachine {
	var vm objects.VirtualMachine
	err := db.GetObj(ctx, ref, &vm)
	if err != nil {
		return nil
	}
	return &vm
}

func (db *DB) GetAllCluster(ctx context.Context) ([]objects.Cluster, error) {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesCluster.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	var clusters []objects.Cluster
	for redisIter.Next(ctx) {
		var cluster objects.Cluster
		redisKey := redisIter.Val()
		err := db.Get(ctx, objects.ManagedObjectTypesCluster, redisKey, &cluster)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)
	}
	return clusters, nil
}

func (db *DB) GetAllComputeResource(ctx context.Context) ([]objects.ComputeResource, error) {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesComputeResource.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	var objs []objects.ComputeResource
	for redisIter.Next(ctx) {
		var obj objects.ComputeResource
		redisKey := redisIter.Val()
		err := db.Get(ctx, objects.ManagedObjectTypesComputeResource, redisKey, &obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

func (db *DB) GetAllDatacenter(ctx context.Context) ([]objects.Datacenter, error) {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesDatacenter.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	var objs []objects.Datacenter
	for redisIter.Next(ctx) {
		var obj objects.Datacenter
		redisKey := redisIter.Val()
		err := db.Get(ctx, objects.ManagedObjectTypesDatacenter, redisKey, &obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

func (db *DB) GetAllDatastore(ctx context.Context) ([]objects.Datastore, error) {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesDatastore.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	var objs []objects.Datastore
	for redisIter.Next(ctx) {
		var obj objects.Datastore
		redisKey := redisIter.Val()
		err := db.Get(ctx, objects.ManagedObjectTypesDatastore, redisKey, &obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

func (db *DB) GetAllFolder(ctx context.Context) ([]objects.Folder, error) {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesFolder.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	var objs []objects.Folder
	for redisIter.Next(ctx) {
		var obj objects.Folder
		redisKey := redisIter.Val()
		err := db.Get(ctx, objects.ManagedObjectTypesFolder, redisKey, &obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

func (db *DB) GetAllHost(ctx context.Context) ([]objects.Host, error) {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesHost.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	var objs []objects.Host
	for redisIter.Next(ctx) {
		var obj objects.Host
		redisKey := redisIter.Val()
		err := db.Get(ctx, objects.ManagedObjectTypesHost, redisKey, &obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

func (db *DB) GetAllResourcePool(ctx context.Context) ([]objects.ResourcePool, error) {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesResourcePool.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	var objs []objects.ResourcePool
	for redisIter.Next(ctx) {
		var obj objects.ResourcePool
		redisKey := redisIter.Val()
		err := db.Get(ctx, objects.ManagedObjectTypesResourcePool, redisKey, &obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

func (db *DB) GetAllStoragePod(ctx context.Context) ([]objects.StoragePod, error) {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesStoragePod.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	var objs []objects.StoragePod
	for redisIter.Next(ctx) {
		var obj objects.StoragePod
		redisKey := redisIter.Val()
		err := db.Get(ctx, objects.ManagedObjectTypesStoragePod, redisKey, &obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

func (db *DB) GetAllTagSets(ctx context.Context) ([]objects.TagSet, error) {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesTagSet.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	var objs []objects.TagSet
	for redisIter.Next(ctx) {
		var obj objects.TagSet
		redisKey := redisIter.Val()
		err := db.Get(ctx, objects.ManagedObjectTypesTagSet, redisKey, &obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

func (db *DB) GetAllVM(ctx context.Context) ([]objects.VirtualMachine, error) {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesVirtualMachine.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	var objs []objects.VirtualMachine
	for redisIter.Next(ctx) {
		var obj objects.VirtualMachine
		redisKey := redisIter.Val()
		err := db.Get(ctx, objects.ManagedObjectTypesVirtualMachine, redisKey, &obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

func (db *DB) GetAllClusterIter(ctx context.Context) iter.Seq[objects.Cluster] {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesCluster.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	return func(yield func(objects.Cluster) bool) {
		for redisIter.Next(ctx) {
			var cluster objects.Cluster
			redisKey := redisIter.Val()
			db.Get(ctx, objects.ManagedObjectTypesCluster, redisKey, &cluster)
			if !yield(cluster) {
				return
			}
		}
	}
}

func (db *DB) GetAllComputeResourceIter(ctx context.Context) iter.Seq[objects.ComputeResource] {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesComputeResource.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	return func(yield func(objects.ComputeResource) bool) {
		for redisIter.Next(ctx) {
			var cp objects.ComputeResource
			redisKey := redisIter.Val()
			db.Get(ctx, objects.ManagedObjectTypesComputeResource, redisKey, &cp)
			if !yield(cp) {
				return
			}
		}
	}
}

func (db *DB) GetAllDatacenterIter(ctx context.Context) iter.Seq[objects.Datacenter] {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesDatacenter.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	return func(yield func(objects.Datacenter) bool) {
		for redisIter.Next(ctx) {
			var dc objects.Datacenter
			redisKey := redisIter.Val()
			db.Get(ctx, objects.ManagedObjectTypesDatacenter, redisKey, &dc)
			if !yield(dc) {
				return
			}
		}
	}
}

func (db *DB) GetAllDatastoreIter(ctx context.Context) iter.Seq[objects.Datastore] {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesDatastore.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	return func(yield func(objects.Datastore) bool) {
		for redisIter.Next(ctx) {
			var ds objects.Datastore
			redisKey := redisIter.Val()
			db.Get(ctx, objects.ManagedObjectTypesDatastore, redisKey, &ds)
			if !yield(ds) {
				return
			}
		}
	}
}

func (db *DB) GetAllFolderIter(ctx context.Context) iter.Seq[objects.Folder] {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesFolder.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	return func(yield func(objects.Folder) bool) {
		for redisIter.Next(ctx) {
			var folder objects.Folder
			redisKey := redisIter.Val()
			db.Get(ctx, objects.ManagedObjectTypesFolder, redisKey, &folder)
			if !yield(folder) {
				return
			}
		}
	}
}

func (db *DB) GetAllHostIter(ctx context.Context) iter.Seq[objects.Host] {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesHost.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	return func(yield func(objects.Host) bool) {
		for redisIter.Next(ctx) {
			var host objects.Host
			redisKey := redisIter.Val()
			db.Get(ctx, objects.ManagedObjectTypesHost, redisKey, &host)
			if !yield(host) {
				return
			}
		}
	}
}

func (db *DB) GetAllStoragePodIter(ctx context.Context) iter.Seq[objects.StoragePod] {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesStoragePod.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	return func(yield func(objects.StoragePod) bool) {
		for redisIter.Next(ctx) {
			var spod objects.StoragePod
			redisKey := redisIter.Val()
			db.Get(ctx, objects.ManagedObjectTypesStoragePod, redisKey, &spod)
			if !yield(spod) {
				return
			}
		}
	}
}

func (db *DB) GetAllResourcePoolIter(ctx context.Context) iter.Seq[objects.ResourcePool] {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesResourcePool.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	return func(yield func(objects.ResourcePool) bool) {
		for redisIter.Next(ctx) {
			var rpool objects.ResourcePool
			redisKey := redisIter.Val()
			db.Get(ctx, objects.ManagedObjectTypesResourcePool, redisKey, &rpool)
			if !yield(rpool) {
				return
			}
		}
	}
}

func (db *DB) GetAllTagSetsIter(ctx context.Context) iter.Seq[objects.TagSet] {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesTagSet.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	return func(yield func(objects.TagSet) bool) {
		for redisIter.Next(ctx) {
			var tagSet objects.TagSet
			redisKey := redisIter.Val()
			db.Get(ctx, objects.ManagedObjectTypesTagSet, redisKey, &tagSet)
			if !yield(tagSet) {
				return
			}
		}
	}
}

func (db *DB) GetAllVMIter(ctx context.Context) iter.Seq[objects.VirtualMachine] {
	db.Connect(ctx)
	match := fmt.Sprintf("%s:*", objects.ManagedObjectTypesVirtualMachine.String())
	redisIter := db.client.Scan(ctx, 0, match, 0).Iterator()
	return func(yield func(objects.VirtualMachine) bool) {
		for redisIter.Next(ctx) {
			var vm objects.VirtualMachine
			redisKey := redisIter.Val()
			db.Get(ctx, objects.ManagedObjectTypesVirtualMachine, redisKey, &vm)
			if !yield(vm) {
				return
			}
		}
	}
}

func (db *DB) GetAllHostRefs(ctx context.Context) []objects.ManagedObjectReference {
	result := []objects.ManagedObjectReference{}
	for host := range db.GetAllHostIter(ctx) {
		result = append(result, host.Self)
	}
	return result
}

func (db *DB) GetAllVMRefs(ctx context.Context) []objects.ManagedObjectReference {
	result := []objects.ManagedObjectReference{}
	for vm := range db.GetAllVMIter(ctx) {
		result = append(result, vm.Self)
	}
	return result
}

func (db *DB) GetAllClusterRefs(ctx context.Context) []objects.ManagedObjectReference {
	result := []objects.ManagedObjectReference{}
	for vm := range db.GetAllClusterIter(ctx) {
		result = append(result, vm.Self)
	}
	return result
}

func (db *DB) SetTags(ctx context.Context, tagSet objects.TagSet, ttl time.Duration) error {
	err := db.Set(ctx, objects.ManagedObjectTypesTagSet, tagSet.ObjectRef.ID(), tagSet, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetTags(ctx context.Context, ref objects.ManagedObjectReference) objects.TagSet {
	var tagSet objects.TagSet
	err := db.Get(ctx, objects.ManagedObjectTypesTagSet, ref.ID(), &tagSet)
	if err != nil {
		panic(err)
	}
	return tagSet
}

func (db *DB) GetParentChain(ctx context.Context, ref objects.ManagedObjectReference) objects.ParentChain {
	return db.walkParentChain(ctx, ref, objects.ParentChain{
		DC:           "",
		Cluster:      "",
		SPOD:         "",
		ResourcePool: "",
		Chain:        []string{},
	})
}

func (db *DB) walkParentChain(ctx context.Context, ref objects.ManagedObjectReference, chain objects.ParentChain) objects.ParentChain {
	chain.Chain = append(chain.Chain, fmt.Sprintf("%s:%s", ref.Type, ref.Value))
	if ref.Type == objects.ManagedObjectTypesCluster {
		cluster := db.GetCluster(ctx, ref)
		if cluster != nil {
			chain.Cluster = cluster.Name
			if cluster.Parent != nil {
				return db.walkParentChain(ctx, *cluster.Parent, chain)
			}
			return chain
		}
	} else if ref.Type == objects.ManagedObjectTypesComputeResource {
		cr := db.GetComputeResource(ctx, ref)
		if cr != nil {
			if cr.Parent != nil {
				return db.walkParentChain(ctx, *cr.Parent, chain)
			}
			return chain
		}
	} else if ref.Type == objects.ManagedObjectTypesDatacenter {
		dc := db.GetDatacenter(ctx, ref)
		if dc != nil {
			chain.DC = dc.Name
			if dc.Parent != nil {
				return db.walkParentChain(ctx, *dc.Parent, chain)
			}
			return chain
		}
	} else if ref.Type == objects.ManagedObjectTypesDatastore {
		ds := db.GetDatastore(ctx, ref)
		if ds != nil {
			if ds.Parent != nil {
				return db.walkParentChain(ctx, *ds.Parent, chain)
			}
			return chain
		}
	} else if ref.Type == objects.ManagedObjectTypesFolder {
		folder := db.GetFolder(ctx, ref)
		if folder != nil {
			if folder.Parent != nil {
				return db.walkParentChain(ctx, *folder.Parent, chain)
			}
			return chain
		}
	} else if ref.Type == objects.ManagedObjectTypesHost {
		host := db.GetHost(ctx, ref)
		if host != nil {
			if host.Parent != nil {
				return db.walkParentChain(ctx, *host.Parent, chain)
			}
			return chain
		}
	} else if ref.Type == objects.ManagedObjectTypesStoragePod {
		spod := db.GetStoragePod(ctx, ref)
		if spod != nil {
			chain.SPOD = spod.Name
			if spod.Parent != nil {
				return db.walkParentChain(ctx, *spod.Parent, chain)
			}
			return chain
		}
	} else if ref.Type == objects.ManagedObjectTypesResourcePool {
		rp := db.GetResourcePool(ctx, ref)
		if rp != nil {
			chain.ResourcePool = rp.Name
			if rp.Parent != nil {
				return db.walkParentChain(ctx, *rp.Parent, chain)
			}
			return chain
		}
	} else if ref.Type == objects.ManagedObjectTypesVirtualMachine {
		vm := db.GetVM(ctx, ref)
		if vm != nil {
			if vm.Parent != nil {
				return db.walkParentChain(ctx, *vm.Parent, chain)
			}
			return chain
		}
	}
	return chain
}

func (db *DB) JsonDump(ctx context.Context, refType objects.ManagedObjectTypes) ([]byte, error) {
	switch refType {
	case objects.ManagedObjectTypesCluster:
		clusters, err := db.GetAllCluster(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(clusters, "", "  ")
	case objects.ManagedObjectTypesComputeResource:
		compResources, err := db.GetAllComputeResource(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(compResources, "", "  ")
	case objects.ManagedObjectTypesDatacenter:
		dcs, err := db.GetAllDatacenter(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(dcs, "", "  ")
	case objects.ManagedObjectTypesDatastore:
		dss, err := db.GetAllDatastore(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(dss, "", "  ")
	case objects.ManagedObjectTypesFolder:
		folders, err := db.GetAllFolder(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(folders, "", "  ")
	case objects.ManagedObjectTypesHost:
		hosts, err := db.GetAllHost(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(hosts, "", "  ")
	case objects.ManagedObjectTypesStoragePod:
		spods, err := db.GetAllStoragePod(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(spods, "", "  ")
	case objects.ManagedObjectTypesResourcePool:
		pools, err := db.GetAllResourcePool(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(pools, "", "  ")
	case objects.ManagedObjectTypesTagSet:
		tagSets, err := db.GetAllTagSets(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(tagSets, "", "  ")
	case objects.ManagedObjectTypesVirtualMachine:
		vms, err := db.GetAllVM(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(vms, "", "  ")
	}
	return nil, nil
}
