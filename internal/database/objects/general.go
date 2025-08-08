package objects

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/vmware/govmomi/vim25/types"
)

type ManagedObjectReference struct {
	Type  ManagedObjectTypes `json:"type" redis:"type"`
	Value string             `json:"value" redis:"value"`
}

func (m ManagedObjectReference) MarshalBinary() ([]byte, error) {
	return []byte(fmt.Sprintf("%s:%s", m.Type.String(), m.Value)), nil
}

func (m *ManagedObjectReference) UnmarshalBinary(data []byte) error {
	s := string(data)
	splt := strings.SplitAfterN(s, ":", 2)
	if len(splt) != 2 {
		return fmt.Errorf("failed to parse reference [%s]", s)
	}
	m.Type = ManagedObjectTypes(splt[0])
	m.Value = splt[1]
	return nil
}

func NewManagedObjectReference(t ManagedObjectTypes, v string) ManagedObjectReference {
	return ManagedObjectReference{
		Type:  t,
		Value: v,
	}
}

func NewManagedObjectReferenceFromVMwareRef(moRef types.ManagedObjectReference) ManagedObjectReference {
	switch t := moRef.Type; t {
	case string(types.ManagedObjectTypesClusterComputeResource):
		return NewManagedObjectReference(ManagedObjectTypesCluster, moRef.Value)
	case string(types.ManagedObjectTypesComputeResource):
		return NewManagedObjectReference(ManagedObjectTypesComputeResource, moRef.Value)
	case string(types.ManagedObjectTypesDatacenter):
		return NewManagedObjectReference(ManagedObjectTypesDatacenter, moRef.Value)
	case string(types.ManagedObjectTypesDatastore):
		return NewManagedObjectReference(ManagedObjectTypesDatastore, moRef.Value)
	case string(types.ManagedObjectTypesFolder):
		return NewManagedObjectReference(ManagedObjectTypesFolder, moRef.Value)
	case string(types.ManagedObjectTypesHostSystem):
		return NewManagedObjectReference(ManagedObjectTypesHost, moRef.Value)
	case string(types.ManagedObjectTypesResourcePool):
		return NewManagedObjectReference(ManagedObjectTypesResourcePool, moRef.Value)
	case string(types.ManagedObjectTypesStoragePod):
		return NewManagedObjectReference(ManagedObjectTypesStoragePod, moRef.Value)
	case string(types.ManagedObjectTypesVirtualMachine):
		return NewManagedObjectReference(ManagedObjectTypesVirtualMachine, moRef.Value)
	default:
		panic(fmt.Sprintf("unknown object type %s", t))
	}
}

// Return a sha256 hash of the type and value of the ManagedObjectReference
func (r *ManagedObjectReference) Hash() string {
	h := sha256.New()
	h.Write([]byte(r.Type))
	h.Write([]byte(r.Value))
	return string(h.Sum(nil))
}

func (r *ManagedObjectReference) ToVMwareRef() types.ManagedObjectReference {
	var t string
	switch typ := r.Type; typ {
	case ManagedObjectTypesCluster:
		t = string(types.ManagedObjectTypesClusterComputeResource)
	case ManagedObjectTypesComputeResource:
		t = string(types.ManagedObjectTypesComputeResource)
	case ManagedObjectTypesDatacenter:
		t = string(types.ManagedObjectTypesDatacenter)
	case ManagedObjectTypesDatastore:
		t = string(types.ManagedObjectTypesDatastore)
	case ManagedObjectTypesHost:
		t = string(types.ManagedObjectTypesHostSystem)
	case ManagedObjectTypesResourcePool:
		t = string(types.ManagedObjectTypesResourcePool)
	case ManagedObjectTypesStoragePod:
		t = string(types.ManagedObjectTypesStoragePod)
	case ManagedObjectTypesVirtualMachine:
		t = string(types.ManagedObjectTypesVirtualMachine)
	case ManagedObjectTypesTagSet:
		panic(fmt.Sprintf("Can not convert TagSet to a types.ManagedObjectReference", typ))
	case ManagedObjectTypesTag:
		panic(fmt.Sprintf("Can not convert Tag to a types.ManagedObjectReference", typ))
	default:
		panic(fmt.Sprintf("unknown internal object type [%s]", typ))
	}

	return types.ManagedObjectReference{
		Type:  t,
		Value: r.Value,
	}
}

func (r *ManagedObjectReference) ID() string {
	return string(r.Value)
}
