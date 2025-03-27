package helper

import "sync"

type StartedCheck struct {
	mu      sync.Mutex
	cond    *sync.Cond
	started bool
}

func NewStartedCheck() *StartedCheck {
	c := &StartedCheck{
		started: false,
	}
	c.cond = sync.NewCond(&c.mu)
	return c
}

func (c *StartedCheck) Wait() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for !c.started {
		c.cond.Wait()
	}
}

func (c *StartedCheck) Started() {
	c.mu.Lock()
	c.started = true
	c.mu.Unlock()
	c.cond.Broadcast()
}

func (c *StartedCheck) Stopped() {
	c.mu.Lock()
	c.started = false
	c.mu.Unlock()
}

func (c *StartedCheck) IsStarted() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.started
}
