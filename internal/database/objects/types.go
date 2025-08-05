package objects

type ManagedObjectTypes string
type PerfMetricTypes string

const (
	ManagedObjectTypesCluster         = ManagedObjectTypes("Cluster")
	ManagedObjectTypesComputeResource = ManagedObjectTypes("ComputeResource")
	ManagedObjectTypesDatacenter      = ManagedObjectTypes("Datacenter")
	ManagedObjectTypesDatastore       = ManagedObjectTypes("Datastore")
	ManagedObjectTypesFolder          = ManagedObjectTypes("Folder")
	ManagedObjectTypesHost            = ManagedObjectTypes("Host")
	ManagedObjectTypesResourcePool    = ManagedObjectTypes("ResourcePool")
	ManagedObjectTypesStoragePod      = ManagedObjectTypes("StoragePod")
	ManagedObjectTypesTag             = ManagedObjectTypes("Tag")
	ManagedObjectTypesTagSet          = ManagedObjectTypes("TagSet")
	ManagedObjectTypesVirtualMachine  = ManagedObjectTypes("VirtualMachine")
)

const (
	PerfMetricTypesVirtualMachine = PerfMetricTypes("PerfMetricsVirtualMachine")
	PerfMetricTypesHost           = PerfMetricTypes("PerfMetricsHost")
)

func (t ManagedObjectTypes) String() string {
	return string(t)
}

func (t PerfMetricTypes) String() string {
	return string(t)
}
