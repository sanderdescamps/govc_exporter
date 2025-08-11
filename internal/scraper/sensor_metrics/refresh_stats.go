package sensormetrics

import "time"

type RefreshStats struct {
	ClientWaitTime time.Duration
	QueryTime      time.Duration
}
