package pool

import "sync"

type MultiAccessPool[T any] struct {
	poolObject *T
	available  chan int
	atExit     func() error
	max        int
	blockLock  sync.Mutex
}

func NewMultiAccessPool[T any](item *T, max int, atExit func() error) *MultiAccessPool[T] {
	available := make(chan int, max)
	for i := 0; i < max; i++ {
		available <- i
	}

	multiPool := MultiAccessPool[T]{
		poolObject: item,
		max:        max,
		available:  available,
	}
	multiPool.SetAtExit(atExit)

	return &multiPool
}

func (p *MultiAccessPool[T]) Acquire() (object *T, release func(), err error) {
	for {
		select {
		case id := <-p.available:
			release := func() {
				p.available <- id
			}
			return p.poolObject, release, nil
		}
	}
}

func (p *MultiAccessPool[T]) SetAtExit(f func() error) {
	p.atExit = f
}

func (p *MultiAccessPool[T]) Destroy() error {
	if p.atExit != nil {
		close(p.available)
		return p.atExit()
	}
	return nil
}
