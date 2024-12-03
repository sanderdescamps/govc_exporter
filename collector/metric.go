package collector

type Metric interface {
	GetValue() float64
	GetName() string
	GetLabels() Labels
	SetLabel(key string, value string)
}

type BasicMetric struct {
	name   string
	value  float64
	Labels Labels
}

func NewBasicMetric(name string, value float64, labels Labels) Metric {
	result := &BasicMetric{name: name, value: value, Labels: labels}
	var parse Metric = result
	return parse
}

func (m *BasicMetric) GetName() string {
	return m.name
}

func (m *BasicMetric) GetValue() float64 {
	return m.value
}

func (m *BasicMetric) GetLabels() Labels {
	return m.Labels
}

func (m *BasicMetric) SetLabel(key string, value string) {
	m.Labels = m.Labels.Add(key, value)
}
