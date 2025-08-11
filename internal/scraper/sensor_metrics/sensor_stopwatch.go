package sensormetrics

import "time"

type SensorStopwatch struct {
	stopClient     Stopwatch
	stopQuery      Stopwatch
	ClientWaitTime time.Duration
	QueryTime      time.Duration
}

func NewSensorStopwatch() *SensorStopwatch {
	return &SensorStopwatch{}
}

func (h *SensorStopwatch) Start() {
	if h.stopClient == nil {
		h.stopClient = NewStopwatch()
	}
}

func (h *SensorStopwatch) Mark1() {
	if h.stopClient != nil {
		h.ClientWaitTime = h.stopClient()
	}
	if h.stopQuery == nil {
		h.stopQuery = NewStopwatch()
	}
}

func (h *SensorStopwatch) Finish() {
	if h.stopQuery != nil {
		h.QueryTime = h.stopQuery()
	}
}

func (h *SensorStopwatch) GetStats() RefreshStats {
	return RefreshStats{
		ClientWaitTime: h.ClientWaitTime,
		QueryTime:      h.QueryTime,
	}
}
