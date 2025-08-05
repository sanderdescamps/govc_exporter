package database

import (
	"context"
	"iter"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
)

type Database interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error

	SetCluster(ctx context.Context, cluster objects.Cluster, ttl time.Duration) error
	SetComputeResource(ctx context.Context, compResource objects.ComputeResource, ttl time.Duration) error
	SetDatacenter(ctx context.Context, ds objects.Datacenter, ttl time.Duration) error
	SetDatastore(ctx context.Context, ds objects.Datastore, ttl time.Duration) error
	SetFolder(ctx context.Context, ds objects.Folder, ttl time.Duration) error
	SetHost(ctx context.Context, host objects.Host, ttl time.Duration) error
	SetStoragePod(ctx context.Context, spod objects.StoragePod, ttl time.Duration) error
	SetResourcePool(ctx context.Context, rp objects.ResourcePool, ttl time.Duration) error
	SetVM(ctx context.Context, vm objects.VirtualMachine, ttl time.Duration) error

	GetCluster(ctx context.Context, ref objects.ManagedObjectReference) *objects.Cluster
	GetComputeResource(ctx context.Context, ref objects.ManagedObjectReference) *objects.ComputeResource
	GetDatacenter(ctx context.Context, ref objects.ManagedObjectReference) *objects.Datacenter
	GetDatastore(ctx context.Context, ref objects.ManagedObjectReference) *objects.Datastore
	GetFolder(ctx context.Context, ref objects.ManagedObjectReference) *objects.Folder
	GetHost(ctx context.Context, ref objects.ManagedObjectReference) *objects.Host
	GetStoragePod(ctx context.Context, ref objects.ManagedObjectReference) *objects.StoragePod
	GetResourcePool(ctx context.Context, ref objects.ManagedObjectReference) *objects.ResourcePool
	GetVM(ctx context.Context, ref objects.ManagedObjectReference) *objects.VirtualMachine

	GetAllClusterIter(ctx context.Context) iter.Seq[objects.Cluster]
	GetAllComputeResourceIter(ctx context.Context) iter.Seq[objects.ComputeResource]
	GetAllDatacenterIter(ctx context.Context) iter.Seq[objects.Datacenter]
	GetAllDatastoreIter(ctx context.Context) iter.Seq[objects.Datastore]
	GetAllFolderIter(ctx context.Context) iter.Seq[objects.Folder]
	GetAllHostIter(ctx context.Context) iter.Seq[objects.Host]
	GetAllStoragePodIter(ctx context.Context) iter.Seq[objects.StoragePod]
	GetAllResourcePoolIter(ctx context.Context) iter.Seq[objects.ResourcePool]
	GetAllTagSetsIter(ctx context.Context) iter.Seq[objects.TagSet]
	GetAllVMIter(ctx context.Context) iter.Seq[objects.VirtualMachine]

	GetAllHostRefs(ctx context.Context) []objects.ManagedObjectReference
	GetAllVMRefs(ctx context.Context) []objects.ManagedObjectReference

	SetTags(ctx context.Context, tagSet objects.TagSet, ttl time.Duration) error
	GetTags(ctx context.Context, ref objects.ManagedObjectReference) objects.TagSet

	GetParentChain(ctx context.Context, ref objects.ManagedObjectReference) objects.ParentChain
	JsonDump(ctx context.Context, refType objects.ManagedObjectTypes) ([]byte, error)
}
