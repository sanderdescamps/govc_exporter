package sensormetrics

import (
	"time"
)

type TimeCounter interface {
	Add(time.Duration)
	GetCurrent() time.Duration
}

type AvgTimeCounter struct {
	currentCount int64
	maxWindow    int64
	avg          time.Duration
}

func (c *AvgTimeCounter) Add(val time.Duration) {
	oldCount := c.currentCount
	if c.currentCount+1 < c.maxWindow {
		c.currentCount = c.currentCount + 1
	} else {
		c.currentCount = c.maxWindow
	}

	c.avg = (time.Duration(oldCount)*c.avg + val) / time.Duration(c.currentCount)
}

func (c *AvgTimeCounter) GetCurrent() time.Duration {
	return c.avg
}

type LastTimeCounter struct {
	last time.Duration
}

func (c *LastTimeCounter) Add(val time.Duration) {
	c.last = val
}

func (c *LastTimeCounter) GetCurrent() time.Duration {
	return c.last
}
