package pool

import (
	"context"
	"sync"
)

type ThrottlerPool[T any] struct {
	poolObject   *T
	available    chan int
	size         int
	hijackActive sync.Mutex
}

func NewThrottlerPool[T any](item *T, size int) *ThrottlerPool[T] {
	available := make(chan int, size)
	for i := 0; i < size; i++ {
		available <- i
	}

	pool := ThrottlerPool[T]{
		poolObject: item,
		size:       size,
		available:  available,
	}

	return &pool
}

func (p *ThrottlerPool[T]) Acquire() (*T, func(), error) {
	id := <-p.available
	release := func() {
		p.available <- id
	}
	return p.poolObject, release, nil
}

func (p *ThrottlerPool[T]) AcquireWithContext(ctx context.Context) (*T, func(), error) {
	for {
		select {
		case id := <-p.available:
			release := func() {
				p.available <- id
			}
			return p.poolObject, release, nil
		case <-ctx.Done():
			return nil, func() {}, ctx.Err()
		}
	}
}

// Wait until you can aquire the entire pool
func (p *ThrottlerPool[T]) Drain() (*T, func(), error) {
	p.hijackActive.Lock()
	defer p.hijackActive.Unlock()
	releaseFunc := []func(){}
	for i := 0; i < p.size; i++ {
		_, release, _ := p.Acquire()
		releaseFunc = append(releaseFunc, release)
	}
	return p.poolObject, func() {
		for _, release := range releaseFunc {
			release()
		}
	}, nil
}

// Wait until you can aquire the entire pool
func (p *ThrottlerPool[T]) DrainWithContext(ctx context.Context) (*T, func(), error) {
	p.hijackActive.Lock()
	defer p.hijackActive.Unlock()
	releaseFunc := []func(){}
	for i := 0; i < p.size; i++ {
		_, release, _ := p.Acquire()
		releaseFunc = append(releaseFunc, release)
	}

	if err := ctx.Err(); err != nil {
		for _, release := range releaseFunc {
			release()
		}
		return nil, nil, err
	} else {
		return p.poolObject, func() {
			for _, release := range releaseFunc {
				release()
			}
		}, nil
	}

}
