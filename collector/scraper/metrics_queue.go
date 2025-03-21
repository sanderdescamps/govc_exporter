package scraper

import (
	"github.com/sanderdescamps/govc_exporter/collector/timequeue"
)

type MetricQueue = timequeue.EventQueue[Metric]

func NewMetricQueue() *MetricQueue {
	return timequeue.NewEventQueue[Metric]()
}
