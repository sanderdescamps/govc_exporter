package memory_db

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database"
	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
	"github.com/vmware/govmomi/vim25/json"
)

const CLEANER_INTERVAL_SEC = 5

type DB struct {
	tabels          map[objects.ManagedObjectTypes]*Table
	lock            sync.Mutex
	cleanerStopChan chan bool
}

func NewDB() *DB {
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
	}
	return db.tabels[name]
}

func (db *DB) Connect(ctx context.Context) error {
	db.startCleaner(ctx)
	return nil
}

func (db *DB) Disconnect(ctx context.Context) error {
	db.stopCleaner(ctx)
	db.tabels = nil
	return nil
}

func (db *DB) startCleaner(ctx context.Context) {
	ticker := time.NewTicker(CLEANUP_INTERVAL)
	go func() {
		for {
			select {
			case <-ticker.C:
				go func() {
					for oType := range db.tabels {
						if t := db.Table(oType); t != nil {
							count := t.CleanupExpired()
							if l := database.GetLoggerFromContext(ctx); l != nil && count > 0 {
								l.Info(fmt.Sprintf("removed %d objects from %s table", count, oType.String()))
							}
						}
					}
				}()
			case <-db.cleanerStopChan:
				ticker.Stop()
				return
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (db *DB) stopCleaner(ctx context.Context) {
	db.cleanerStopChan <- true
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

func (db *DB) GetAllCluster(ctx context.Context) ([]objects.Cluster, error) {
	var allObjs []objects.Cluster
	err := db.Table(objects.ManagedObjectTypesCluster).GetAll(&allObjs)
	if err != nil {
		return nil, err
	}
	return allObjs, nil
}

func (db *DB) GetAllComputeResource(ctx context.Context) ([]objects.ComputeResource, error) {
	var allObjs []objects.ComputeResource
	err := db.Table(objects.ManagedObjectTypesComputeResource).GetAll(&allObjs)
	if err != nil {
		return nil, err
	}
	return allObjs, nil
}

func (db *DB) GetAllDatacenter(ctx context.Context) ([]objects.Datacenter, error) {
	var allObjs []objects.Datacenter
	err := db.Table(objects.ManagedObjectTypesDatacenter).GetAll(&allObjs)
	if err != nil {
		return nil, err
	}
	return allObjs, nil
}

func (db *DB) GetAllDatastore(ctx context.Context) ([]objects.Datastore, error) {
	var allObjs []objects.Datastore
	err := db.Table(objects.ManagedObjectTypesDatastore).GetAll(&allObjs)
	if err != nil {
		return nil, err
	}
	return allObjs, nil
}

func (db *DB) GetAllFolder(ctx context.Context) ([]objects.Folder, error) {
	var allObjs []objects.Folder
	err := db.Table(objects.ManagedObjectTypesFolder).GetAll(&allObjs)
	if err != nil {
		return nil, err
	}
	return allObjs, nil
}

func (db *DB) GetAllHost(ctx context.Context) ([]objects.Host, error) {
	var allObjs []objects.Host
	err := db.Table(objects.ManagedObjectTypesHost).GetAll(&allObjs)
	if err != nil {
		return nil, err
	}
	return allObjs, nil
}

func (db *DB) GetAllResourcePool(ctx context.Context) ([]objects.ResourcePool, error) {
	var allObjs []objects.ResourcePool
	err := db.Table(objects.ManagedObjectTypesResourcePool).GetAll(&allObjs)
	if err != nil {
		return nil, err
	}
	return allObjs, nil
}

func (db *DB) GetAllStoragePod(ctx context.Context) ([]objects.StoragePod, error) {
	var allObjs []objects.StoragePod
	err := db.Table(objects.ManagedObjectTypesStoragePod).GetAll(&allObjs)
	if err != nil {
		return nil, err
	}
	return allObjs, nil
}

func (db *DB) GetAllTagSets(ctx context.Context) ([]objects.TagSet, error) {
	var allObjs []objects.TagSet
	err := db.Table(objects.ManagedObjectTypesTagSet).GetAll(&allObjs)
	if err != nil {
		return nil, err
	}
	return allObjs, nil
}

func (db *DB) GetAllVM(ctx context.Context) ([]objects.VirtualMachine, error) {
	var allObjs []objects.VirtualMachine
	err := db.Table(objects.ManagedObjectTypesVirtualMachine).GetAll(&allObjs)
	if err != nil {
		return nil, err
	}
	return allObjs, nil
}

func (db *DB) GetAllHostRefs(ctx context.Context) []objects.ManagedObjectReference {
	hosts, _ := db.GetAllHost(ctx)
	result := []objects.ManagedObjectReference{}
	for _, host := range hosts {
		result = append(result, host.Self)
	}
	return result
}

func (db *DB) GetAllVMRefs(ctx context.Context) []objects.ManagedObjectReference {
	vms, _ := db.GetAllVM(ctx)
	result := []objects.ManagedObjectReference{}
	for _, vm := range vms {
		result = append(result, vm.Self)
	}
	return result
}

func (db *DB) GetAllClusterRefs(ctx context.Context) []objects.ManagedObjectReference {
	clusters, _ := db.GetAllCluster(ctx)
	result := []objects.ManagedObjectReference{}
	for _, cluster := range clusters {
		result = append(result, cluster.Self)
	}
	return result
}

func (db *DB) GetAllDatacenterRefs(ctx context.Context) []objects.ManagedObjectReference {
	dcs, _ := db.GetAllDatacenter(ctx)
	result := []objects.ManagedObjectReference{}
	for _, dc := range dcs {
		result = append(result, dc.Self)
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
		clusters, err := db.GetAllCluster(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(clusters, "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesComputeResource {
		compResources, err := db.GetAllComputeResource(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(compResources, "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesDatacenter {
		dcs, err := db.GetAllDatacenter(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(dcs, "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesDatastore {
		dss, err := db.GetAllDatastore(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(dss, "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesFolder {
		folders, err := db.GetAllFolder(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(folders, "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesHost {
		hosts, err := db.GetAllHost(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(hosts, "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesStoragePod {
		spods, err := db.GetAllStoragePod(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(spods, "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesResourcePool {
		pools, err := db.GetAllResourcePool(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(pools, "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesTagSet {
		tagSets, err := db.GetAllTagSets(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(tagSets, "", "  ")
	} else if db.HasTable(refType) && refType == objects.ManagedObjectTypesVirtualMachine {
		vms, err := db.GetAllVM(ctx)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(vms, "", "  ")
	}
	return nil, nil
}
