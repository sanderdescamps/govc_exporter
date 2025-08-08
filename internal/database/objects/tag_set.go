package objects

type TagSet struct {
	ObjectRef ManagedObjectReference `json:"object_ref" redis:"object_ref"`
	Tags      map[string]string      `json:"tags" redis:"tags"`
}

func (o *TagSet) GetTag(catName string) string {
	if tag, ok := o.Tags[catName]; ok {
		return tag
	}
	return ""
}
