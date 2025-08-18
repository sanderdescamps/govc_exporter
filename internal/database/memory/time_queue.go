package memory_db

import (
	"iter"
	"sort"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/database/objects"
)

type MetricItem struct {
	Metric *objects.Metric
	Expire time.Time
}

type TimeQueueTable struct {
	lock  sync.RWMutex
	queue []*MetricItem
}

func NewTimeQueueTable() *TimeQueueTable {
	return &TimeQueueTable{
		queue: []*MetricItem{},
	}
}

func (q *TimeQueueTable) Len() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.queue)
}

func (q *TimeQueueTable) Add(ttl time.Duration, objs ...objects.Metric) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.insert(ttl, objs...)
}

func (q *TimeQueueTable) insert(ttl time.Duration, objs ...objects.Metric) {
	if q.queue == nil {
		q.queue = []*MetricItem{}
	}
	for _, obj := range objs {
		expireTime := obj.Timestamp.Add(ttl)
		if time.Now().After(expireTime) {
			continue
		}
		q.queue = append(q.queue, nil)
		i := sort.Search(len(q.queue), func(i int) bool { return q.queue[i] == nil || expireTime.Before((q.queue[i]).Expire) })
		copy(q.queue[i+1:], q.queue[i:])
		q.queue[i] = &MetricItem{
			Metric: &obj,
			Expire: expireTime,
		}
	}
}

func (q *TimeQueueTable) pop() *objects.Metric {
	if len(q.queue) > 0 {
		first, new_queue := q.queue[0], q.queue[1:]
		q.queue = new_queue
		return first.Metric
	}
	return nil
}

// Empty the queue and return a popper function with all the items
func (q *TimeQueueTable) popAll() []*objects.Metric {
	var dump *[]*MetricItem = &q.queue
	q.queue = []*MetricItem{}

	result := []*objects.Metric{}
	for _, m := range *dump {
		result = append(result, m.Metric)
	}

	return result
}

func (q *TimeQueueTable) Pop() *objects.Metric {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.pop()
}

// Empty the queue and return a popper function with all the items
func (q *TimeQueueTable) PopAll() []*objects.Metric {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.popAll()
}

func (q *TimeQueueTable) popAllIter() iter.Seq[objects.Metric] {
	dump := q.queue
	q.queue = nil
	return func(yield func(objects.Metric) bool) {
		for _, m := range dump {
			if m.Metric != nil && !yield(*m.Metric) {
				return
			}
		}
	}
}

func (q *TimeQueueTable) PopAllIter() iter.Seq[objects.Metric] {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.popAllIter()
}

func (q *TimeQueueTable) CleanupExpired() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	if len(q.queue) == 0 {
		return 0
	}

	i := sort.Search(len(q.queue), func(i int) bool { return time.Now().Before((q.queue[i]).Expire) })

	older, younger := q.queue[0:i], q.queue[i:]
	q.queue = younger
	return len(older)
}

func (q *TimeQueueTable) JsonDump() ([]byte, error) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return json.MarshalIndent(q.queue, "", "  ")
}
