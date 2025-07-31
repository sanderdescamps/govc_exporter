package objects

import (
	"crypto/sha256"
)

type ManagedObjectReference struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func NewManagedObjectReference(t string, v string) ManagedObjectReference {
	return ManagedObjectReference{
		Type:  t,
		Value: v,
	}
}

// Return a sha256 hash of the type and value of the ManagedObjectReference
func (r *ManagedObjectReference) Hash() string {
	h := sha256.New()
	h.Write([]byte(r.Type))
	h.Write([]byte(r.Value))
	return string(h.Sum(nil))
}
