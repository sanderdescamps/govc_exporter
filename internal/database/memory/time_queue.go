package memory_db

import (
	"sort"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database"
)

type TimeQueueTable struct {
	lock  sync.RWMutex
	queue []database.TimeItem
}

func NewTimeQueueTable() *TimeQueueTable {
	return &TimeQueueTable{
		queue: []database.TimeItem{},
	}
}

func (q *TimeQueueTable) Len() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.queue)
}

func (q *TimeQueueTable) Add(objs ...database.TimeItem) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.insert(objs...)
}

func (q *TimeQueueTable) insert(objs ...database.TimeItem) {
	if q.queue == nil {
		q.queue = []database.TimeItem{}
	}
	for _, obj := range objs {
		if obj == nil {
			continue
		}
		q.queue = append(q.queue, nil)
		i := sort.Search(len(q.queue), func(i int) bool { return q.queue[i] == nil || (obj).TimeKey().Before((q.queue[i]).TimeKey()) })
		copy(q.queue[i+1:], q.queue[i:])
		q.queue[i] = obj
	}
}

func (q *TimeQueueTable) pop() database.TimeItem {
	if len(q.queue) > 0 {
		first, new_queue := q.queue[0], q.queue[1:]
		q.queue = new_queue
		return first
	}
	return nil
}

// Empty the queue and return a popper function with all the items
func (q *TimeQueueTable) popAll() []database.TimeItem {
	dump := q.queue
	q.queue = nil

	return dump
}

func (q *TimeQueueTable) Pop() database.TimeItem {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.pop()
}

// Empty the queue and return a popper function with all the items
func (q *TimeQueueTable) PopAll() []database.TimeItem {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.popAll()
}

func (q *TimeQueueTable) PopOlderOrEqualThan(t time.Time) []database.TimeItem {
	if len(q.queue) == 0 {
		return nil
	}

	i := func() int {
		q.lock.RLock()
		defer q.lock.RUnlock()
		return sort.Search(len(q.queue), func(i int) bool { return t.Before((q.queue[i]).TimeKey()) })
	}()

	q.lock.Lock()
	defer q.lock.Unlock()

	older, younger := q.queue[0:i], q.queue[i:]
	q.queue = younger

	return older
}

// func (q *TimeQueueTable) PopAllIter() iter.Seq[database.TimeItem] {
// 	return func(yield func(database.TimeItem) bool) {
// 		items := q.PopAll()
// 		for _, v := range items {
// 			if !yield(v) {
// 				return
// 			}
// 		}
// 	}
// }
