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

	SetCluster(ctx context.Context, cluster *objects.Cluster, ttl time.Duration) error
	SetComputeResource(ctx context.Context, compResource *objects.ComputeResource, ttl time.Duration) error
	SetDatastore(ctx context.Context, ds *objects.Datastore, ttl time.Duration) error
	SetHost(ctx context.Context, host *objects.Host, ttl time.Duration) error
	SetStoragePod(ctx context.Context, spod *objects.StoragePod, ttl time.Duration) error
	SetResourcePool(ctx context.Context, rp *objects.ResourcePool, ttl time.Duration) error
	SetVM(ctx context.Context, vm *objects.VirtualMachine, ttl time.Duration) error

	GetCluster(ctx context.Context, ref objects.ManagedObjectReference) *objects.Cluster
	GetComputeResource(ctx context.Context, ref objects.ManagedObjectReference) *objects.ComputeResource
	GetDatastore(ctx context.Context, ref objects.ManagedObjectReference) *objects.Datastore
	GetHost(ctx context.Context, ref objects.ManagedObjectReference) *objects.Host
	GetStoragePod(ctx context.Context, ref objects.ManagedObjectReference) *objects.StoragePod
	GetResourcePool(ctx context.Context, ref objects.ManagedObjectReference) *objects.ResourcePool
	GetVM(ctx context.Context, ref objects.ManagedObjectReference) *objects.VirtualMachine

	GetAllClusterIter(ctx context.Context) iter.Seq2[*objects.Cluster, error]
	GetAllComputeResourceIter(ctx context.Context) iter.Seq2[*objects.ComputeResource, error]
	GetAllDatastoreIter(ctx context.Context) iter.Seq2[*objects.Datastore, error]
	GetAllHostIter(ctx context.Context) iter.Seq2[*objects.Host, error]
	GetAllStoragePodIter(ctx context.Context) iter.Seq2[*objects.StoragePod, error]
	GetAllResourcePoolIter(ctx context.Context) iter.Seq2[*objects.ResourcePool, error]
	GetAllVMIter(ctx context.Context) iter.Seq2[*objects.VirtualMachine, error]

	GetAllHostRefs(ctx context.Context) []objects.ManagedObjectReference
	GetAllVMRefs(ctx context.Context) []objects.ManagedObjectReference

	SetObjectTag(ctx context.Context, tag objects.ObjectTag, ttl time.Duration) error
	GetObjectTags(ref objects.ManagedObjectReference) *objects.ObjectTag

	GetManagedEntity(ctx context.Context, ref objects.ManagedObjectReference) *objects.ManagedEntity
	GetParentChain(ctx context.Context, ref objects.ManagedObjectReference) ParentChain
}

type ParentChain struct {
	DC           string
	Cluster      string
	ResourcePool string
	SPOD         string
	Chain        []string
}
