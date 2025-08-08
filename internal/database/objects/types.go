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

func (m ManagedObjectTypes) MarshalBinary() ([]byte, error) {
	return []byte(m.String()), nil
}

func (m *ManagedObjectTypes) UnmarshalBinary(data []byte) error {
	*m = ManagedObjectTypes(string(data))
	return nil
}

func (t PerfMetricTypes) String() string {
	return string(t)
}

func (m PerfMetricTypes) MarshalBinary() ([]byte, error) {
	return []byte(m.String()), nil
}

func (m *PerfMetricTypes) UnmarshalBinary(data []byte) error {
	*m = PerfMetricTypes(string(data))
	return nil
}
