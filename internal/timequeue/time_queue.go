package timequeue

import (
	"encoding/json"
	"iter"
	"sort"
	"sync"
	"time"
)

type TimeQueue[T Event] struct {
	lock  sync.RWMutex
	queue []*T
}

func NewTimeQueue[T Event]() *TimeQueue[T] {
	return &TimeQueue[T]{
		queue: []*T{},
	}
}

func (q *TimeQueue[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Queue []*T `json:"queue"`
	}{
		Queue: q.queue,
	})
}

func (q *TimeQueue[T]) Len() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.queue)
}

func (q *TimeQueue[T]) Add(objs ...*T) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.insert(objs...)
}

func (q *TimeQueue[T]) insert(objs ...*T) {
	if q.queue == nil {
		q.queue = []*T{}
	}
	for _, obj := range objs {
		if obj == nil {
			continue
		}
		q.queue = append(q.queue, nil)
		i := sort.Search(len(q.queue), func(i int) bool { return q.queue[i] == nil || (*obj).Timestamp().Before((*q.queue[i]).Timestamp()) })
		copy(q.queue[i+1:], q.queue[i:])
		q.queue[i] = obj
	}
}

func (q *TimeQueue[T]) pop() *T {
	if len(q.queue) > 0 {
		first, new_queue := q.queue[0], q.queue[1:]
		q.queue = new_queue
		return first
	}
	return nil
}

// Empty the queue and return a popper function with all the items
func (q *TimeQueue[T]) popAll() []*T {
	dump := q.queue
	q.queue = nil

	return dump
}

func (q *TimeQueue[T]) Pop() *T {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.pop()
}

// Empty the queue and return a popper function with all the items
func (q *TimeQueue[T]) PopAll() []*T {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.popAll()
}

func (q *TimeQueue[T]) PopOlderOrEqualThan(t time.Time) []*T {
	if len(q.queue) == 0 {
		return nil
	}

	i := func() int {
		q.lock.RLock()
		defer q.lock.RUnlock()
		return sort.Search(len(q.queue), func(i int) bool { return t.Before((*q.queue[i]).Timestamp()) })
	}()

	q.lock.Lock()
	defer q.lock.Unlock()

	older, younger := q.queue[0:i], q.queue[i:]
	q.queue = younger

	return older
}

func (q *TimeQueue[T]) PopAllItems() iter.Seq[*T] {
	return func(yield func(*T) bool) {
		items := q.PopAll()
		for _, v := range items {
			if !yield(v) {
				return
			}
		}
	}
}
