package pool

import "sync"

// type PoolObject interface {
// 	Init() error
// 	ReInit() error
// 	Destroy() error
// 	Healthy() error
// }

type ThrottlerPool[T any] struct {
	poolObject   *T
	available    chan int
	size         int
	hijackActive sync.Mutex
	wg           sync.WaitGroup
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

func (p *ThrottlerPool[T]) Acquire() (object *T, release func(), err error) {
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

// Wait until you can aquire the entire pool
func (p *ThrottlerPool[T]) Drain() (object *T, release func(), err error) {
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
