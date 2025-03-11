package scraper

import "time"

type SensorMetric struct {
	Name           string
	QueryTime      time.Duration
	ClientWaitTime time.Duration
	Status         bool
}

func (m *SensorMetric) TotalRefreshTime() time.Duration {
	return m.ClientWaitTime + m.QueryTime
}

func (m *SensorMetric) RollingMean(metric SensorMetric, window int64) SensorMetric {
	return SensorMetric{
		Name:           metric.Name,
		Status:         metric.Status,
		ClientWaitTime: time.Duration(m.ClientWaitTime.Nanoseconds() + (metric.ClientWaitTime.Nanoseconds()-m.ClientWaitTime.Nanoseconds())/window),
		QueryTime:      time.Duration(m.QueryTime.Nanoseconds() + (metric.QueryTime.Nanoseconds()-m.QueryTime.Nanoseconds())/window),
	}
}
