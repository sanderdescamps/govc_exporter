package collector

type Pair[T any] struct {
	Key, Value T
}

type Labels []Pair[string]

// type Labels struct {
// 	labels []Pair[string]
// }

func (b Labels) Add(k string, v string) Labels {
	b = append(b, Pair[string]{Key: k, Value: v})
	return b
}

func (b Labels) GetKeys() []string {
	keys := []string{}
	for _, p := range b {
		keys = append(keys, p.Key)
	}
	return keys
}

func (b Labels) GetValues() []string {
	values := []string{}
	for _, p := range b {
		values = append(values, p.Value)
	}
	return values
}
