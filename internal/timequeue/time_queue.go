package timequeue

import (
	"encoding/json"
	"sort"
	"sync"
	"time"
)

type QueueObj[T any] struct {
	Timestamp time.Time `json:"time_stamp"`
	Obj       *T        `json:"obj"`
}

type TimeQueue[T any] struct {
	lock  sync.RWMutex
	queue []*QueueObj[T] `json:"queue"`
}

func (q *TimeQueue[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Queue []*QueueObj[T] `json:"queue"`
	}{
		Queue: q.queue,
	})
}

func (q *TimeQueue[T]) Len() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.queue)
}

func (q *TimeQueue[T]) insert(obj *QueueObj[T]) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.queue == nil {
		q.queue = []*QueueObj[T]{}
	}
	q.queue = append(q.queue, nil)
	i := sort.Search(len(q.queue), func(i int) bool { return q.queue[i] == nil || obj.Timestamp.Before(q.queue[i].Timestamp) })
	copy(q.queue[i+1:], q.queue[i:])
	q.queue[i] = obj
}

func (q *TimeQueue[T]) pop() (time.Time, *T) {
	if len(q.queue) > 0 {
		first, new_queue := q.queue[0], q.queue[1:]
		q.queue = new_queue
		return first.Timestamp, first.Obj
	}
	return time.Time{}, nil
}

func (q *TimeQueue[T]) Add(timestamp time.Time, obj *T) {
	q.insert(&QueueObj[T]{
		Timestamp: timestamp,
		Obj:       obj,
	})
}

func (q *TimeQueue[T]) Pop() (time.Time, *T) {
	return q.pop()
}

// Empty the queue and return a popper function with all the items
func (q *TimeQueue[T]) PopAll() []*QueueObj[T] {
	q.lock.Lock()
	defer q.lock.Unlock()
	queueCopy := q.queue
	q.queue = nil

	return queueCopy
}

func (q *TimeQueue[T]) PopOlderOrEqualThan(t time.Time) []*QueueObj[T] {
	if len(q.queue) == 0 {
		return nil
	}

	i := func() int {
		q.lock.RLock()
		defer q.lock.RUnlock()
		return sort.Search(len(q.queue), func(i int) bool { return t.Before(q.queue[i].Timestamp) })
	}()

	q.lock.Lock()
	defer q.lock.Unlock()

	older, younger := q.queue[0:i], q.queue[i:]
	q.queue = younger

	return older
}
