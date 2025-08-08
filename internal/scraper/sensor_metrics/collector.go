package sensormetrics

type MetricsCollector interface {
	GetLatestMetrics() []SensorMetric
}

type SensorMetricsCollector struct {
	clientWaitTime TimeCounter
	queryTime      TimeCounter
}

func NewAvgSensorMetricsCollector(maxWindow int64) *SensorMetricsCollector {
	return &SensorMetricsCollector{
		clientWaitTime: &AvgTimeCounter{maxWindow: maxWindow},
		queryTime:      &AvgTimeCounter{maxWindow: maxWindow},
	}
}

func NewLastSensorMetricsCollector() *SensorMetricsCollector {
	return &SensorMetricsCollector{
		clientWaitTime: &LastTimeCounter{},
		queryTime:      &LastTimeCounter{},
	}
}

func (h *SensorMetricsCollector) UploadStats(stats RefreshStats) {
	if stats.ClientWaitTime != 0 {
		h.clientWaitTime.Add(stats.ClientWaitTime)
	}
	if stats.QueryTime != 0 {
		h.queryTime.Add(stats.QueryTime)
	}
}

func (h *SensorMetricsCollector) ComposeMetrics(sensorName string) []SensorMetric {
	result := []SensorMetric{}
	result = append(result, SensorMetric{
		Sensor:     sensorName,
		MetricName: "client_wait_time",
		Value:      float64(h.clientWaitTime.GetCurrent().Nanoseconds()),
		Unit:       "nanoseconds",
	})
	result = append(result, SensorMetric{
		Sensor:     sensorName,
		MetricName: "query_time",
		Value:      float64(h.queryTime.GetCurrent().Nanoseconds()),
		Unit:       "nanoseconds",
	})
	return result
}
