package objects

import "time"

// Metric

type Metric struct {
	Ref       ManagedObjectReference `json:"ref"`
	Name      string                 `json:"name"`
	Unit      string                 `json:"unit"`
	Instance  string                 `json:"instance"`
	Value     float64                `json:"value"`
	Timestamp time.Time              `json:"time_stamp"`
}

func (m *Metric) TimeKey() time.Time {
	return m.Timestamp
}
