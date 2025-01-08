package scraper

import "time"

// type SensorMetricType string

// const (
// 	SensorMetricClientWaitTime SensorMetricType = "ClientWaitTime"
// 	SensorMetricQueryTime      SensorMetricType = "QueryTime"
// 	SensorMetricRefreshTime    SensorMetricType = "RefreshTime"
// 	SensorMetricStatus         SensorMetricType = "Status"
// )

// type SensorMetricCollector struct {
// 	lock    sync.RWMutex
// 	metrics SensorMetrics
// }

// type SensorMetrics map[SensorMetricType]interface{}

// func (m *SensorMetricCollector) Update(t SensorMetricType, val interface{}) {
// 	m.lock.Lock()
// 	defer m.lock.Unlock()
// 	m.metrics[t] = val
// }

// func (m *SensorMetricCollector) Export() SensorMetrics {
// 	return m.metrics
// }

// func (m *SensorMetrics) Get(t SensorMetricType) interface{} {
// 	if val, ok := m.Get(t)[t]; ok {
// 		return val
// 	} else {
// 		return nil
// 	}
// }

// func (m *SensorMetrics) TotalRefreshTime() time.Duration {
// 	return m.ClientWaitTime() + m.QueryTime()
// }

// func (m *SensorMetrics) QueryTime() time.Duration {
// 	val := m.Get(SensorMetricQueryTime)
// 	if val != nil {
// 		return val.(time.Duration)
// 	}
// 	return 0
// }

// func (m *SensorMetrics) ClientWaitTime() time.Duration {
// 	val := m.Get(SensorMetricClientWaitTime)
// 	if val != nil {
// 		return val.(time.Duration)
// 	}
// 	return 0
// }

// func (m *SensorMetrics) Status() bool {
// 	val := m.Get(SensorMetricStatus)
// 	if val != nil {
// 		return val.(bool)
// 	}
// 	return false
// }

type SensorMetrics struct {
	Name           string
	QueryTime      time.Duration
	ClientWaitTime time.Duration
	Status         bool
}

func (m *SensorMetrics) TotalRefreshTime() time.Duration {
	return m.ClientWaitTime + m.QueryTime
}

type SensorMetric[T string | time.Duration] struct {
	Sensor string
	Name   string
	Value  T
}
