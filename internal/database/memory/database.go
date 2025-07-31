package memory_db

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database"
	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type DB struct {
	tabels     map[string]*Table
	timeQueues map[string]map[string]*TimeQueueTable
}

func NewDB() *DB {
	return &DB{
		tabels: make(map[string]*Table),
	}
}

func NewDBWithCleaner() *DB {
	return &DB{
		tabels: make(map[string]*Table),
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

func (db *DB) HasTable(name string) bool {
	if _, ok := db.tabels[name]; ok {
		return true
	}
	return false
}

func (db *DB) Table(name string) *Table {
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

func (db *DB) SetObj(ctx context.Context, key string, path string, obj interface{}, ttl time.Duration) error {
	err := db.Table(path).SetWithTTL(key, obj, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetCluster(ctx context.Context, cluster *mo.ClusterComputeResource, ttl time.Duration) error {
	err := db.SetObj(ctx, cluster.Self.Value, string(types.ManagedObjectTypesClusterComputeResource), cluster, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetComputeResource(ctx context.Context, compResource *mo.ComputeResource, ttl time.Duration) error {
	err := db.SetObj(ctx, compResource.Self.Value, string(types.ManagedObjectTypesComputeResource), compResource, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetDatastore(ctx context.Context, ds *mo.Datastore, ttl time.Duration) error {
	err := db.SetObj(ctx, ds.Self.Value, string(types.ManagedObjectTypesDatastore), ds, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetHost(ctx context.Context, host *mo.HostSystem, ttl time.Duration) error {
	err := db.SetObj(ctx, host.Self.Value, string(types.ManagedObjectTypesHostSystem), host, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetStoragePod(ctx context.Context, spod *mo.StoragePod, ttl time.Duration) error {
	err := db.SetObj(ctx, spod.Self.Value, string(types.ManagedObjectTypesStoragePod), spod, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetResourcePool(ctx context.Context, rp *mo.ResourcePool, ttl time.Duration) error {
	err := db.SetObj(ctx, rp.Self.Value, string(types.ManagedObjectTypesResourcePool), rp, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetVM(ctx context.Context, vm *mo.VirtualMachine, ttl time.Duration) error {
	err := db.SetObj(ctx, vm.Self.Value, string(types.ManagedObjectTypesVirtualMachine), vm, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetCluster(ctx context.Context, ref types.ManagedObjectReference) *mo.ClusterComputeResource {
	var cluster mo.ClusterComputeResource
	err := db.Table(string(types.ManagedObjectTypesClusterComputeResource)).Get(ref.Value, &cluster)
	if err != nil {
		return nil
	}
	return &cluster
}

func (db *DB) GetComputeResource(ctx context.Context, ref types.ManagedObjectReference) *mo.ComputeResource {
	var compResource mo.ComputeResource
	err := db.Table(string(types.ManagedObjectTypesComputeResource)).Get(ref.Value, &compResource)
	if err != nil {
		return nil
	}
	return &compResource
}

func (db *DB) GetDatastore(ctx context.Context, ref types.ManagedObjectReference) *mo.Datastore {
	var ds mo.Datastore
	err := db.Table(string(types.ManagedObjectTypesDatastore)).Get(ref.Value, &ds)
	if err != nil {
		return nil
	}
	return &ds
}

func (db *DB) GetHost(ctx context.Context, ref types.ManagedObjectReference) *mo.HostSystem {
	var host mo.HostSystem
	err := db.Table(string(types.ManagedObjectTypesHostSystem)).Get(ref.Value, &host)
	if err != nil {
		return nil
	}
	return &host
}

func (db *DB) GetStoragePod(ctx context.Context, ref types.ManagedObjectReference) *mo.StoragePod {
	var spod mo.StoragePod
	err := db.Table(string(types.ManagedObjectTypesStoragePod)).Get(ref.Value, &spod)
	if err != nil {
		return nil
	}
	return &spod
}

func (db *DB) GetResourcePool(ctx context.Context, ref types.ManagedObjectReference) *mo.ResourcePool {
	var rp mo.ResourcePool
	err := db.Table(string(types.ManagedObjectTypesResourcePool)).Get(ref.Value, &rp)
	if err != nil {
		return nil
	}
	return &rp
}

func (db *DB) GetVM(ctx context.Context, ref types.ManagedObjectReference) *mo.VirtualMachine {
	var vm mo.VirtualMachine
	err := db.Table(string(types.ManagedObjectTypesVirtualMachine)).Get(ref.Value, &vm)
	if err != nil {
		return nil
	}
	return &vm
}

func (db *DB) GetAllClusterIter(ctx context.Context) iter.Seq2[*mo.ClusterComputeResource, error] {
	return func(yield func(*mo.ClusterComputeResource, error) bool) {
		var clusters []mo.ClusterComputeResource
		db.Table(string(types.ManagedObjectTypesClusterComputeResource)).GetAll(&clusters)
		for _, cluster := range clusters {
			if !yield(&cluster, nil) {
				return
			}
		}
	}
}

func (db *DB) GetAllComputeResourceIter(ctx context.Context) iter.Seq2[*mo.ComputeResource, error] {
	return func(yield func(*mo.ComputeResource, error) bool) {
		var compResources []mo.ComputeResource
		db.Table(string(types.ManagedObjectTypesComputeResource)).GetAll(&compResources)
		for _, compResource := range compResources {
			if !yield(&compResource, nil) {
				return
			}
		}
	}
}

func (db *DB) GetAllDatastoreIter(ctx context.Context) iter.Seq2[*mo.Datastore, error] {
	return func(yield func(*mo.Datastore, error) bool) {
		var datastores []mo.Datastore
		db.Table(string(types.ManagedObjectTypesDatastore)).GetAll(&datastores)
		for _, ds := range datastores {
			if !yield(&ds, nil) {
				return
			}
		}
	}
}

func (db *DB) GetAllHostIter(ctx context.Context) iter.Seq2[*mo.HostSystem, error] {
	return func(yield func(*mo.HostSystem, error) bool) {
		var hosts []mo.HostSystem
		db.Table(string(types.ManagedObjectTypesHostSystem)).GetAll(&hosts)
		for _, host := range hosts {
			if !yield(&host, nil) {
				return
			}
		}
	}
}

func (db *DB) GetAllStoragePodIter(ctx context.Context) iter.Seq2[*mo.StoragePod, error] {
	return func(yield func(*mo.StoragePod, error) bool) {
		var spods []mo.StoragePod
		db.Table(string(types.ManagedObjectTypesStoragePod)).GetAll(&spods)
		for _, spod := range spods {
			if !yield(&spod, nil) {
				return
			}
		}
	}
}

func (db *DB) GetAllResourcePoolIter(ctx context.Context) iter.Seq2[*mo.ResourcePool, error] {
	return func(yield func(*mo.ResourcePool, error) bool) {
		var pools []mo.ResourcePool
		db.Table(string(types.ManagedObjectTypesResourcePool)).GetAll(&pools)
		for _, pool := range pools {
			if !yield(&pool, nil) {
				return
			}
		}
	}
}

func (db *DB) GetAllVMIter(ctx context.Context) iter.Seq2[*mo.VirtualMachine, error] {
	return func(yield func(*mo.VirtualMachine, error) bool) {
		var vms []mo.VirtualMachine
		db.Table(string(types.ManagedObjectTypesVirtualMachine)).GetAll(&vms)
		for _, vm := range vms {
			if !yield(&vm, nil) {
				return
			}
		}
	}
}

func (db *DB) GetAllHostRefs(ctx context.Context) []types.ManagedObjectReference {
	result := []types.ManagedObjectReference{}
	for host := range db.GetAllHostIter(ctx) {
		result = append(result, host.Self)
	}
	return result
}

func (db *DB) GetAllVMRefs(ctx context.Context) []types.ManagedObjectReference {
	result := []types.ManagedObjectReference{}
	for vm := range db.GetAllVMIter(ctx) {
		result = append(result, vm.Self)
	}
	return result
}

func (db *DB) SetCategories(ctx context.Context, cats []tags.Category, ttl time.Duration) error {
	for _, cat := range cats {
		err := db.Table("category").SetWithTTL(cat.ID, cat, ttl)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) SetObjectTag(ctx context.Context, tag objects.ObjectTag, ttl time.Duration) error {
	err := db.Table("tags").SetWithTTL(tag.ObjectRef.Hash(), tag, ttl)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetObjectTags(ref objects.ManagedObjectReference) *objects.ObjectTag {
	var tag objects.ObjectTag
	err := db.Table("tags").Get(ref.Hash(), tag)
	if err != nil {
		return nil
	}
	return &tag
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

func (db *DB) GetManagedEntity(ctx context.Context, ref types.ManagedObjectReference) *mo.ManagedEntity {
	if db.HasTable(string(types.ManagedObjectTypesClusterComputeResource)) && ref.Type == string(types.ManagedObjectTypesClusterComputeResource) {
		cluster := db.GetCluster(ctx, ref)
		if cluster != nil {
			return &cluster.ManagedEntity
		}
	} else if db.HasTable(string(types.ManagedObjectTypesComputeResource)) && ref.Type == string(types.ManagedObjectTypesClusterComputeResource) {
		cr := db.GetComputeResource(ctx, ref)
		if cr != nil {
			return &cr.ManagedEntity
		}
	} else if db.HasTable(string(types.ManagedObjectTypesDatastore)) && ref.Type == string(types.ManagedObjectTypesDatastore) {
		ds := db.GetDatastore(ctx, ref)
		if ds != nil {
			return &ds.ManagedEntity
		}
	} else if db.HasTable(string(types.ManagedObjectTypesHostSystem)) && ref.Type == string(types.ManagedObjectTypesHostSystem) {
		host := db.GetHost(ctx, ref)
		if host != nil {
			return &host.ManagedEntity
		}
	} else if db.HasTable(string(types.ManagedObjectTypesStoragePod)) && ref.Type == string(types.ManagedObjectTypesStoragePod) {
		spod := db.GetStoragePod(ctx, ref)
		if spod != nil {
			return &spod.ManagedEntity
		}
	} else if db.HasTable(string(types.ManagedObjectTypesResourcePool)) && ref.Type == string(types.ManagedObjectTypesResourcePool) {
		rp := db.GetResourcePool(ctx, ref)
		if rp != nil {
			return &rp.ManagedEntity
		}
	} else if db.HasTable(string(types.ManagedObjectTypesVirtualMachine)) && ref.Type == string(types.ManagedObjectTypesVirtualMachine) {
		vm := db.GetVM(ctx, ref)
		if vm != nil {
			return &vm.ManagedEntity
		}
	}
	return nil
}

func (db *DB) GetParentChain(ctx context.Context, ref types.ManagedObjectReference) database.ParentChain {
	return db.walkParentChain(ctx, ref, database.ParentChain{
		DC:           "",
		Cluster:      "",
		SPOD:         "",
		ResourcePool: "",
		Chain:        []string{fmt.Sprintf("%s:%s", ref.Type, ref.Value)},
	})
}

func (db *DB) walkParentChain(ctx context.Context, ref types.ManagedObjectReference, chain database.ParentChain) database.ParentChain {
	if db.HasTable(string(types.ManagedObjectTypesClusterComputeResource)) && ref.Type == string(types.ManagedObjectTypesClusterComputeResource) {
		cluster := db.GetCluster(ctx, ref)
		if cluster != nil {
			chain.Cluster = cluster.Name
			if cluster.Parent != nil {
				return db.walkParentChain(ctx, *cluster.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(string(types.ManagedObjectTypesComputeResource)) && ref.Type == string(types.ManagedObjectTypesClusterComputeResource) {
		cr := db.GetComputeResource(ctx, ref)
		if cr != nil {
			if cr.Parent != nil {
				return db.walkParentChain(ctx, *cr.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(string(types.ManagedObjectTypesDatastore)) && ref.Type == string(types.ManagedObjectTypesDatastore) {
		ds := db.GetDatastore(ctx, ref)
		if ds != nil {
			chain.Cluster = ds.Name
			if ds.Parent != nil {
				return db.walkParentChain(ctx, *ds.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(string(types.ManagedObjectTypesHostSystem)) && ref.Type == string(types.ManagedObjectTypesHostSystem) {
		host := db.GetHost(ctx, ref)
		if host != nil {
			chain.Cluster = host.Name
			if host.Parent != nil {
				return db.walkParentChain(ctx, *host.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(string(types.ManagedObjectTypesStoragePod)) && ref.Type == string(types.ManagedObjectTypesStoragePod) {
		spod := db.GetStoragePod(ctx, ref)
		if spod != nil {
			chain.SPOD = spod.Name
			if spod.Parent != nil {
				return db.walkParentChain(ctx, *spod.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(string(types.ManagedObjectTypesResourcePool)) && ref.Type == string(types.ManagedObjectTypesResourcePool) {
		rp := db.GetResourcePool(ctx, ref)
		if rp != nil {
			chain.ResourcePool = rp.Name
			if rp.Parent != nil {
				return db.walkParentChain(ctx, *rp.Parent, chain)
			}
			return chain
		}
	} else if db.HasTable(string(types.ManagedObjectTypesVirtualMachine)) && ref.Type == string(types.ManagedObjectTypesVirtualMachine) {
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
