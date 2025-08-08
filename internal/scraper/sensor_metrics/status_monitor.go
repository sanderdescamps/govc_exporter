package sensormetrics

import "sync"

type StatusMonitor struct {
	lock       sync.RWMutex
	failed     bool
	failCount  int64
	totalCount int64
}

func NewStatusMonitor() *StatusMonitor {
	return &StatusMonitor{
		failed: false,
	}
}

func (m *StatusMonitor) Fail() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.totalCount += 1
	m.failCount += 1
	m.failed = true
}

func (m *StatusMonitor) Success() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.totalCount += 1
	m.failed = false
}

func (m *StatusMonitor) Reset() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.failed = false
}

func (m *StatusMonitor) StatusFailed() bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.failed
}

func (m *StatusMonitor) StatusFailedFloat64() float64 {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if m.failed {
		return 1.0
	}
	return 0.0
}

func (m *StatusMonitor) FailRate() float64 {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return float64(m.failCount) / float64(m.totalCount)
}
