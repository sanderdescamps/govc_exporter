package sensormetrics

import "time"

type Stopwatch func() time.Duration

func NewStopwatch() Stopwatch {
	now := time.Now()
	return func() time.Duration {
		return time.Since(now)
	}
}
