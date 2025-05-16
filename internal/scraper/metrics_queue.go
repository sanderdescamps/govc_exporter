package scraper

import (
	"github.com/sanderdescamps/govc_exporter/internal/timequeue"
)

type MetricQueue = timequeue.TimeQueue[Metric]

func NewMetricQueue() *MetricQueue {
	return timequeue.NewTimeQueue[Metric]()
}
