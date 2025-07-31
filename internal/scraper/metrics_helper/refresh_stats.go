package metricshelper

import "time"

type RefreshStats struct {
	ClientWaitTime time.Duration
	QueryTime      time.Duration
	Failed         bool
}
