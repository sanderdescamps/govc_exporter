package metricshelper

import (
	"errors"
	"sync"
	"time"
)

type Stopwatch struct {
	running sync.Mutex
	start   time.Time
	lastest time.Duration
}

func (s *Stopwatch) Start() (stop func(), err error) {
	if s.running.TryLock() {
		s.start = time.Now()
		return func() {
			s.lastest = time.Since(s.start)
			s.running.Unlock()
		}, nil
	} else {
		return nil, errors.New("Stopwatch already running")
	}
}

func (s *Stopwatch) End() error {
	if s.running.TryLock() {
		s.running.Unlock()
		return errors.New("Stopwatch not running")
	} else {
		s.lastest = time.Since(s.start)
		s.start = time.Time{}
		s.running.Unlock()
	}
	return nil
}

func (s *Stopwatch) IsRunning() bool {
	if s.running.TryLock() {
		s.running.Unlock()
		return false
	} else {
		return true
	}
}

func (s *Stopwatch) Latest() time.Duration {
	return s.lastest
}
