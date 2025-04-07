package helper

import "sync"

type StartedCheck struct {
	mu      sync.Mutex
	cond    *sync.Cond
	started bool
}

func NewStartedCheck() *StartedCheck {
	var mu = new(sync.Mutex)
	var cond = sync.NewCond(mu)
	c := &StartedCheck{
		started: false,
		mu:      *mu,
		cond:    cond,
	}
	c.cond = sync.NewCond(&c.mu)
	return c
}

// Wait until the Started method is called
func (c *StartedCheck) Wait() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for !c.started {
		c.cond.Wait()
	}
}

func (c *StartedCheck) Started() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.started = true
	c.cond.Broadcast()
}

func (c *StartedCheck) Stopped() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.started = false
	c.cond.Broadcast()
}

func (c *StartedCheck) IsStarted() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.started
}
