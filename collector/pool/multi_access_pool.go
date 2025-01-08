package pool

import (
	"sync"
)

type MultiAccessPool[T any] struct {
	poolObject *T
	counter    int
	max        int
	lock       sync.Mutex
	atExit     func() error
}

func NewMultiAccessPoolFromCreatorFunc[T any](creator func() *T, max int) Pool[T] {
	multiPool := MultiAccessPool[T]{
		poolObject: creator(),
		max:        max,
	}

	return &multiPool
}

func NewMultiAccessPool[T any](item *T, max int, atExit func() error) Pool[T] {
	multiPool := MultiAccessPool[T]{
		poolObject: item,
		max:        max,
		counter:    0,
	}
	multiPool.SetAtExit(atExit)

	return &multiPool
}

func (p *MultiAccessPool[T]) Acquire() (*T, int) {
	for {
		raised := func() bool {
			p.lock.Lock()
			defer p.lock.Unlock()
			if p.counter < p.max {
				p.counter += 1
				return true
			}
			return false
		}()
		if raised {
			return p.poolObject, 1
		}
	}
}

func (p *MultiAccessPool[T]) Release(id int) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.counter -= 1
}

func (p *MultiAccessPool[T]) SetAtExit(f func() error) {
	p.atExit = f
}

func (p *MultiAccessPool[T]) Close() error {
	if p.atExit != nil {
		return p.atExit()
	}
	return nil
}
