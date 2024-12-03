package collector

import (
	"sync"
	"time"
)

type Sensor struct {
	// parentHub       *SensorHub
	Namespace       string
	LastRefresh     time.Time
	refreshPeriod   time.Duration
	refreshFunction func() ([]Metric, error)
	metrics         []Metric
	lock            sync.Mutex
}

func NewSensor(namespace string, f func() ([]Metric, error), refreshPeriod time.Duration) *Sensor {
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

func (s *Sensor) GetLatestMetrics() ([]Metric, error) {
	if s.Expired() {
		err := s.Refresh()
		if err != nil {
			return nil, err
		}
	}

	return s.GetMetrics(), nil
}

func (s *Sensor) Expired() bool {
	return s.metrics == nil || time.Now().After(s.LastRefresh.Add(s.refreshPeriod))
}
