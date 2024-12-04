package collector

import (
	"sync"
	"time"

	"github.com/vmware/govmomi/vim25/mo"
)

type Sensor struct {
	// parentHub       *SensorHub
	Namespace       string
	LastRefresh     time.Time
	refreshPeriod   time.Duration
	refreshFunction func() ([]mo.ManagedEntity, error)
	lock            sync.Mutex
}

func NewSensor(namespace string, f func() ([]mo.ManagedEntity, error), refreshPeriod time.Duration) *Sensor {
	s := Sensor{
		Namespace:       namespace,
		refreshFunction: f,
		LastRefresh:     time.Date(0, 0, 0, 0, 0, 0, 0, time.Local),
		refreshPeriod:   refreshPeriod,
		lock:            sync.Mutex{},
	}
	return &s
}

func (s *Sensor) Refresh() error {
	metrics, err := s.refreshFunction()
	if err != nil {
		return err
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.metrics = metrics
	return nil
}

func (s *Sensor) GetMetrics() []Metric {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.metrics != nil {
		return s.metrics
	}
	return []Metric{}
}