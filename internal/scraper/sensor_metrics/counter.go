package sensormetrics

import (
	"math"
)

type Counter interface {
	Add(float64)
	GetCurrent() float64
}

type AvgCounter struct {
	currentCount float64
	maxWindow    float64
	avg          float64
}

func (c *AvgCounter) Add(val float64) {
	oldCount := c.currentCount
	c.currentCount = math.Min(c.currentCount+1, c.maxWindow)
	c.avg = (c.avg*oldCount + val) / c.currentCount
}

func (c *AvgCounter) GetCurrent() float64 {
	return c.avg
}

type LastCounter struct {
	last float64
}

func (c *LastCounter) Add(val float64) {
	c.last = val
}

func (c *LastCounter) GetCurrent() float64 {
	return c.last
}
