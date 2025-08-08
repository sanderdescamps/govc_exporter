package objects

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// Metric

type Metric struct {
	Ref       ManagedObjectReference `json:"ref" redis:"-"`
	Name      string                 `json:"name" redis:"name"`
	Unit      string                 `json:"unit" redis:"unit"`
	Instance  string                 `json:"instance" redis:"instance"`
	Value     float64                `json:"value" redis:"value"`
	Timestamp time.Time              `json:"time_stamp" redis:"time_stamp"`
}

func (m *Metric) TimeKey() time.Time {
	return m.Timestamp
}

func (m *Metric) Hash() string {
	h := sha256.New()
	h.Write([]byte(m.Ref.Type))
	h.Write([]byte(m.Ref.Value))
	h.Write([]byte(m.Unit))
	h.Write([]byte(m.Instance))
	h.Write([]byte(m.Name))
	h.Write([]byte(m.Timestamp.String()))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// func (m Metric) MarshalBinary() ([]byte, error) {
// 	return json.Marshal(m)
// }

func (m *Metric) UnmarshalBinary(data []byte) error {

	s := &Metric{}
	err := json.Unmarshal(data, s)
	if err != nil {
		return err
	}
	*m = *s
	return nil
}
