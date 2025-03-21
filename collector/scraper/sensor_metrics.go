package scraper

import (
	"sync"
	"time"
)

type BaseSensorMetric struct {
	Name          string
	Unit          string
	Value         float64
	LastUpdate    time.Time
	AverageWindow int
	init          bool
}

type SensorMetric struct {
	BaseSensorMetric
	lock sync.RWMutex
}

func NewSensorMetric(name string, unit string, avgWindow int) *SensorMetric {
	return &SensorMetric{
		BaseSensorMetric: BaseSensorMetric{
			Name:          name,
			Unit:          unit,
			AverageWindow: avgWindow,
			LastUpdate:    time.Now(),
			Value:         0,
		},
	}
}

func (m *SensorMetric) Update(value float64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.AverageWindow < 1 {
		m.Value = value
	} else if !m.init {
		m.Value = value
		m.init = true
	} else {
		m.Value = m.Value + (value-m.Value)/float64(m.AverageWindow)
	}
	m.LastUpdate = time.Now()
}

// time Duration

type SensorMetricDuration struct {
	SensorMetric
}

func NewSensorMetricDuration(name string, avgWindow int) *SensorMetricDuration {
	return &SensorMetricDuration{
		SensorMetric: *NewSensorMetric(name, "nanoseconds", avgWindow),
	}
}

func (m *SensorMetricDuration) Update(t time.Duration) {
	value := float64(t.Nanoseconds())
	m.SensorMetric.Update(value)
}

// Boolean

type SensorMetricStatus struct {
	SensorMetric
	allValues         []bool
	keepStatusHistory bool
}

func NewSensorMetricStatus(name string, keepStatusHistory bool) *SensorMetricStatus {
	return &SensorMetricStatus{
		SensorMetric:      *NewSensorMetric(name, "boolean", 0),
		allValues:         []bool{},
		keepStatusHistory: keepStatusHistory,
	}
}

func (m *SensorMetricStatus) Update(b bool) {
	var value bool = b
	if m.keepStatusHistory {
		value = func() bool {
			m.SensorMetric.lock.Lock()
			defer m.SensorMetric.lock.Unlock()
			m.allValues = append(m.allValues, b)
			return AllTrue(m.allValues)
		}()
	}
	if value {
		m.SensorMetric.Update(1)
	} else {
		m.SensorMetric.Update(0)
	}
}

func (m *SensorMetricStatus) Reset() {
	m.SensorMetric.lock.Lock()
	defer m.SensorMetric.lock.Unlock()
	m.Value = 0.0
	m.allValues = []bool{}
}

func (m *SensorMetricStatus) Fail() {
	m.Update(false)
}

func (m *SensorMetricStatus) Success() {
	m.Update(true)
}
