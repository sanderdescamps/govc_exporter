package scraper

import (
	"github.com/sanderdescamps/govc_exporter/internal/timequeue"
)

type MetricQueue = timequeue.EventQueue[Metric]

func NewMetricQueue() *MetricQueue {
	return timequeue.NewEventQueue[Metric]()
}
