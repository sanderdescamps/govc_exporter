package memory_db

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"slices"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/vmware/govmomi/vim25/json"
)

type DB struct {
	tabels map[objects.ManagedObjectTypes]*Table
	lock   sync.Mutex
}

func NewDB() *DB {
	return &DB{
		tabels: make(map[objects.ManagedObjectTypes]*Table),
	}
}

func NewDBWithCleaner() *DB {
	return &DB{
		tabels: make(map[objects.ManagedObjectTypes]*Table),
	}
}

// func (db *DB) HasTable(name string) bool {
// 	if _, ok := db.tabels[name]; ok {
// 		return true
// 	}
// 	return false
// }

// func (db *DB) CreateTable(name string) error {
// 	if db.HasTable(name) {
// 		db.tabels[name] = NewTable()
// 		return nil
// 	}
// 	return errors.New("table already exists")
// }

func (db *DB) HasTable(name objects.ManagedObjectTypes) bool {
	db.lock.Lock()
	defer db.lock.Unlock()
	if _, ok := db.tabels[name]; ok {
		return true
	}
	return false
}

func (db *DB) Table(name objects.ManagedObjectTypes) *Table {
	db.lock.Lock()
	defer db.lock.Unlock()
	if _, ok := db.tabels[name]; !ok {
		db.tabels[name] = NewTable()
		db.tabels[name].StartCleaner()
	}
	return db.tabels[name]
}

func (db *DB) Connect(ctx context.Context) error {
	for _, t := range db.tabels {
		t.StartCleaner()
	}
	return nil
}

func (db *DB) Disconnect(ctx context.Context) error {
	db.tabels = nil
	return nil
}

func (db *DB) SetObj(ctx context.Context, key string, table objects.ManagedObjectTypes, obj interface{}, ttl time.Duration) error {
	err := db.Table(table).SetWithTTL(key, obj, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetCluster(ctx context.Context, cluster objects.Cluster, ttl time.Duration) error {
	err := db.SetObj(ctx, cluster.Self.Value, objects.ManagedObjectTypesCluster, cluster, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetComputeResource(ctx context.Context, compResource objects.ComputeResource, ttl time.Duration) error {
	err := db.SetObj(ctx, compResource.Self.Value, objects.ManagedObjectTypesComputeResource, compResource, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetDatacenter(ctx context.Context, dc objects.Datacenter, ttl time.Duration) error {
	err := db.SetObj(ctx, dc.Self.Value, objects.ManagedObjectTypesDatacenter, dc, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetDatastore(ctx context.Context, ds objects.Datastore, ttl time.Duration) error {
	err := db.SetObj(ctx, ds.Self.Value, objects.ManagedObjectTypesDatastore, ds, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetFolder(ctx context.Context, f objects.Folder, ttl time.Duration) error {
	err := db.SetObj(ctx, f.Self.Value, objects.ManagedObjectTypesFolder, f, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetHost(ctx context.Context, host objects.Host, ttl time.Duration) error {
	err := db.SetObj(ctx, host.Self.Value, objects.ManagedObjectTypesHost, host, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetStoragePod(ctx context.Context, spod objects.StoragePod, ttl time.Duration) error {
	err := db.SetObj(ctx, spod.Self.Value, objects.ManagedObjectTypesStoragePod, spod, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetResourcePool(ctx context.Context, rp objects.ResourcePool, ttl time.Duration) error {
	err := db.SetObj(ctx, rp.Self.Value, objects.ManagedObjectTypesResourcePool, rp, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetVM(ctx context.Context, vm objects.VirtualMachine, ttl time.Duration) error {
	err := db.SetObj(ctx, vm.Self.Value, objects.ManagedObjectTypesVirtualMachine, vm, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetCluster(ctx context.Context, ref objects.ManagedObjectReference) *objects.Cluster {
	var cluster objects.Cluster
	err := db.Table(objects.ManagedObjectTypesCluster).Get(ref.Value, &cluster)
	if err != nil {
		return nil
	}
	return &cluster
}

func (db *DB) GetComputeResource(ctx context.Context, ref objects.ManagedObjectReference) *objects.ComputeResource {
	var compResource objects.ComputeResource
	err := db.Table(objects.ManagedObjectTypesComputeResource).Get(ref.Value, &compResource)
	if err != nil {
		return nil
	}
	return &compResource
}

func (db *DB) GetDatacenter(ctx context.Context, ref objects.ManagedObjectReference) *objects.Datacenter {
	var dc objects.Datacenter
	err := db.Table(objects.ManagedObjectTypesDatacenter).Get(ref.Value, &dc)
	if err != nil {
		return nil
	}
	return &dc
}

func (db *DB) GetDatastore(ctx context.Context, ref objects.ManagedObjectReference) *objects.Datastore {
	var ds objects.Datastore
	err := db.Table(objects.ManagedObjectTypesDatastore).Get(ref.Value, &ds)
	if err != nil {
		return nil
	}
	return &ds
}

func (db *DB) GetFolder(ctx context.Context, ref objects.ManagedObjectReference) *objects.Folder {
	var folder objects.Folder
	err := db.Table(objects.ManagedObjectTypesFolder).Get(ref.Value, &folder)
	if err != nil {
		return nil
	}
	return &folder
}

func (db *DB) GetHost(ctx context.Context, ref objects.ManagedObjectReference) *objects.Host {
	var host objects.Host
	err := db.Table(objects.ManagedObjectTypesHost).Get(ref.Value, &host)
	if err != nil {
		return nil
	}
	return &host
}

func (db *DB) GetStoragePod(ctx context.Context, ref objects.ManagedObjectReference) *objects.StoragePod {
	var spod objects.StoragePod
	err := db.Table(objects.ManagedObjectTypesStoragePod).Get(ref.Value, &spod)
	if err != nil {
		return nil
	}
	return &spod
}

func (db *DB) GetResourcePool(ctx context.Context, ref objects.ManagedObjectReference) *objects.ResourcePool {
	var rp objects.ResourcePool
	err := db.Table(objects.ManagedObjectTypesResourcePool).Get(ref.Value, &rp)
	if err != nil {
		return nil
	}
	return &rp
}

func (db *DB) GetVM(ctx context.Context, ref objects.ManagedObjectReference) *objects.VirtualMachine {
	var vm objects.VirtualMachine
	err := db.Table(objects.ManagedObjectTypesVirtualMachine).Get(ref.Value, &vm)
	if err != nil {
		return nil
	}
	return &vm
}

func (db *DB) GetAllClusterIter(ctx context.Context) iter.Seq[objects.Cluster] {
	clusterIter := db.Table(objects.ManagedObjectTypesCluster).GetAllIter()
	return func(yield func(objects.Cluster) bool) {
		for item := range clusterIter {
			if cluster, ok := item.(objects.Cluster); ok {
				if !yield(cluster) {
					return
				}
			}
		}
	}
}

func (db *DB) GetAllComputeResourceIter(ctx context.Context) iter.Seq[objects.ComputeResource] {
	compResourceIter := db.Table(objects.ManagedObjectTypesComputeResource).GetAllIter()
	return func(yield func(objects.ComputeResource) bool) {
		for item := range compResourceIter {
			if cp, ok := item.(objects.ComputeResource); ok {
				if !yield(cp) {
					return
				}
			}
		}
	}
}

func (db *DB) GetAllDatacenterIter(ctx context.Context) iter.Seq[objects.Datacenter] {
	dcIter := db.Table(objects.ManagedObjectTypesDatacenter).GetAllIter()
	return func(yield func(objects.Datacenter) bool) {
		for item := range dcIter {
			if dc, ok := item.(objects.Datacenter); ok {
				if !yield(dc) {
					return
				}
			}
		}
	}
}

func (db *DB) GetAllDatastoreIter(ctx context.Context) iter.Seq[objects.Datastore] {
	datastoreIter := db.Table(objects.ManagedObjectTypesDatastore).GetAllIter()
	return func(yield func(objects.Datastore) bool) {
		for item := range datastoreIter {
			if datastore, ok := item.(objects.Datastore); ok {
				if !yield(datastore) {
					return
				}
			}
		}
	}
}

func (db *DB) GetAllFolderIter(ctx context.Context) iter.Seq[objects.Folder] {
	folderIter := db.Table(objects.ManagedObjectTypesFolder).GetAllIter()
	return func(yield func(objects.Folder) bool) {
		for item := range folderIter {
			if folder, ok := item.(objects.Folder); ok {
				if !yield(folder) {
					return
				}
			}
		}
	}
}

func (db *DB) GetAllHostIter(ctx context.Context) iter.Seq[objects.Host] {
	hostIter := db.Table(objects.ManagedObjectTypesHost).GetAllIter()
	return func(yield func(objects.Host) bool) {
		for item := range hostIter {
			if host, ok := item.(objects.Host); ok {
				if !yield(host) {
					return
				}
			}
		}
	}
}

func (db *DB) GetAllStoragePodIter(ctx context.Context) iter.Seq[objects.StoragePod] {
	spodIter := db.Table(objects.ManagedObjectTypesStoragePod).GetAllIter()
	return func(yield func(objects.StoragePod) bool) {
		for item := range spodIter {
			if spod, ok := item.(objects.StoragePod); ok {
				if !yield(spod) {
					return
				}
			}
		}
	}
}

func (db *DB) GetAllResourcePoolIter(ctx context.Context) iter.Seq[objects.ResourcePool] {
	rpoolIter := db.Table(objects.ManagedObjectTypesResourcePool).GetAllIter()
	return func(yield func(objects.ResourcePool) bool) {
		for item := range rpoolIter {
			if rpool, ok := item.(objects.ResourcePool); ok {
				if !yield(rpool) {
					return
				}
			}
		}
	}
}

func (db *DB) GetAllTagSetsIter(ctx context.Context) iter.Seq[objects.TagSet] {
	tagIter := db.Table(objects.ManagedObjectTypesTagSet).GetAllIter()
	return func(yield func(objects.TagSet) bool) {
		for item := range tagIter {
			if rpool, ok := item.(objects.TagSet); ok {
				if !yield(rpool) {
					return
				}
			}
		}
	}
}

func (db *DB) GetAllVMIter(ctx context.Context) iter.Seq[objects.VirtualMachine] {
	vmIter := db.Table(objects.ManagedObjectTypesVirtualMachine).GetAllIter()
	return func(yield func(objects.VirtualMachine) bool) {
		for item := range vmIter {
			if vm, ok := item.(objects.VirtualMachine); ok {
				if !yield(vm) {
					return
				}
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

func (db *DB) SetTags(ctx context.Context, tagSet objects.TagSet, ttl time.Duration) error {
	err := db.Table(objects.ManagedObjectTypesTagSet).SetWithTTL(tagSet.ObjectRef.Hash(), tagSet, ttl)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetTags(ctx context.Context, ref objects.ManagedObjectReference) objects.TagSet {
	var tag objects.TagSet
	err := db.Table(objects.ManagedObjectTypesTagSet).Get(ref.Hash(), &tag)
	if errors.Is(err, ErrKeyNotFound) {
		return objects.TagSet{}
	} else if err != nil {
		panic(err)
	}
	return tag
}

// func (db *DB) GetCategoryID(catName string) (string, error) {
// 	var cats []tags.Category
// 	if err := db.Table("category").FindByProp("Name", catName, &cats); err != nil {
// 		return "", err
// 	} else if len(cats) < 1 {
// 		return "", errors.New("category not found")
// 	}

// 	return cats[0].ID, nil
// }

// func (db *DB) GetTagForRef(ref types.ManagedObjectReference, catID string) *tags.Tag {
// 	path := "tags/" + catID

// 	var tag tags.Tag
// 	err := db.Table(path).Get(ref.Value, tag)
// 	if err != nil {
// 		return nil
// 	}
// 	return &tag
// }

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
	if db.HasTable(ref.Type) && ref.Type == objects.ManagedObjectTypesCluster {
		cluster := db.GetCluster(ctx, ref)
		if cluster != nil {
			chain.Cluster = cluster.Name
			if cluster.Parent != nil {
				return db.walkParentChain(ctx, *cluster.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(ref.Type) && ref.Type == objects.ManagedObjectTypesComputeResource {
		cr := db.GetComputeResource(ctx, ref)
		if cr != nil {
			if cr.Parent != nil {
				return db.walkParentChain(ctx, *cr.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(ref.Type) && ref.Type == objects.ManagedObjectTypesDatacenter {
		dc := db.GetDatacenter(ctx, ref)
		if dc != nil {
			chain.DC = dc.Name
			if dc.Parent != nil {
				return db.walkParentChain(ctx, *dc.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(ref.Type) && ref.Type == objects.ManagedObjectTypesDatastore {
		ds := db.GetDatastore(ctx, ref)
		if ds != nil {
			if ds.Parent != nil {
				return db.walkParentChain(ctx, *ds.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(ref.Type) && ref.Type == objects.ManagedObjectTypesFolder {
		folder := db.GetFolder(ctx, ref)
		if folder != nil {
			if folder.Parent != nil {
				return db.walkParentChain(ctx, *folder.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(ref.Type) && ref.Type == objects.ManagedObjectTypesHost {
		host := db.GetHost(ctx, ref)
		if host != nil {
			if host.Parent != nil {
				return db.walkParentChain(ctx, *host.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(ref.Type) && ref.Type == objects.ManagedObjectTypesStoragePod {
		spod := db.GetStoragePod(ctx, ref)
		if spod != nil {
			chain.SPOD = spod.Name
			if spod.Parent != nil {
				return db.walkParentChain(ctx, *spod.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(ref.Type) && ref.Type == objects.ManagedObjectTypesResourcePool {
		rp := db.GetResourcePool(ctx, ref)
		if rp != nil {
			chain.ResourcePool = rp.Name
			if rp.Parent != nil {
				return db.walkParentChain(ctx, *rp.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(ref.Type) && ref.Type == objects.ManagedObjectTypesVirtualMachine {
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
	if db.HasTable(refType) && refType == objects.ManagedObjectTypesCluster {
		return json.MarshalIndent(slices.Collect(db.GetAllClusterIter(ctx)), "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesComputeResource {
		return json.MarshalIndent(slices.Collect(db.GetAllComputeResourceIter(ctx)), "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesDatacenter {
		return json.MarshalIndent(slices.Collect(db.GetAllDatacenterIter(ctx)), "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesDatastore {
		return json.MarshalIndent(slices.Collect(db.GetAllDatastoreIter(ctx)), "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesFolder {
		return json.MarshalIndent(slices.Collect(db.GetAllFolderIter(ctx)), "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesHost {
		return json.MarshalIndent(slices.Collect(db.GetAllHostIter(ctx)), "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesStoragePod {
		return json.MarshalIndent(slices.Collect(db.GetAllStoragePodIter(ctx)), "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesResourcePool {
		return json.MarshalIndent(slices.Collect(db.GetAllResourcePoolIter(ctx)), "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesTagSet {
		return json.MarshalIndent(slices.Collect(db.GetAllTagSetsIter(ctx)), "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesVirtualMachine {
		return json.MarshalIndent(slices.Collect(db.GetAllVMIter(ctx)), "", "  ")
	}
	return nil, fmt.Errorf("unsupported object type %s", refType.String())
}
