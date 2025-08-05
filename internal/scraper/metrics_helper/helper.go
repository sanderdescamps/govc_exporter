package metricshelper

type MetricHelper interface {
	GetLatestMetrics() []SensorMetric
}

type MetricHelperDefault struct {
	sensorName     string
	clientWaitTime Stopwatch
	queryTime      Stopwatch
	status         Status
}

func NewMetricHelperDefault(sensorName string) *MetricHelperDefault {
	return &MetricHelperDefault{
		sensorName:     sensorName,
		clientWaitTime: Stopwatch{},
		queryTime:      Stopwatch{},
		status:         Status{},
	}
}

func (h *MetricHelperDefault) Start() {
	h.clientWaitTime.Start()
}

func (h *MetricHelperDefault) Mark1() {
	h.clientWaitTime.End()
	h.queryTime.Start()
}

func (h *MetricHelperDefault) Finish(success bool) {
	h.queryTime.End()
	h.status.Set(success)
}

func (h *MetricHelperDefault) Fail() {
	if h.clientWaitTime.IsRunning() {
		h.clientWaitTime.End()
	} else if h.queryTime.IsRunning() {
		h.queryTime.End()
	}

	h.status.Fail()
}

func (h *MetricHelperDefault) LoadStats(stats RefreshStats) {
	h.clientWaitTime.lastest = stats.ClientWaitTime
	h.queryTime.lastest = stats.QueryTime
	h.status.Set(stats.Failed)
}

func (h *MetricHelperDefault) GetLatestMetrics() []SensorMetric {
	result := []SensorMetric{}
	result = append(result, SensorMetric{
		Sensor:     h.sensorName,
		MetricName: "client_wait_time",
		Value:      float64(h.clientWaitTime.Latest().Nanoseconds()),
		Unit:       "nanoseconds",
	})
	result = append(result, SensorMetric{
		Sensor:     h.sensorName,
		MetricName: "query_time",
		Value:      float64(h.queryTime.Latest().Nanoseconds()),
		Unit:       "nanoseconds",
	})
	result = append(result, SensorMetric{
		Sensor:     h.sensorName,
		MetricName: "status",
		Value:      h.status.GetFloat64(),
		Unit:       "boolean",
	})
	result = append(result, SensorMetric{
		Sensor:     h.sensorName,
		MetricName: "success_rate",
		Value:      float64(h.status.SuccessRate()),
		Unit:       "boolean",
	})
	result = append(result, SensorMetric{
		Sensor:     h.sensorName,
		MetricName: "enabled",
		Value:      1.0,
		Unit:       "boolean",
	})
	return result
}
