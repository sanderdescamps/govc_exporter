package objects

type ObjectTag struct {
	ObjectRef ManagedObjectReference `json:"object_ref"`
	Tags      map[string]string      `json:"tags"`
}
