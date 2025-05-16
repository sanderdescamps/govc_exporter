package timequeue

import (
	"time"
)

type EventQueue[T interface {
	Timestamp() time.Time
}] struct {
	TimeQueue[T]
}

func NewEventQueue[T interface {
	Timestamp() time.Time
}]() *EventQueue[T] {
	return &EventQueue[T]{
		TimeQueue: TimeQueue[T]{},
	}
}

// func (q *EventQueue[T]) Add(e T) {
// 	timestamp := e.Timestamp()
// 	q.Insert(&QueueObj[T]{
// 		Timestamp: timestamp,
// 		Obj:       &e,
// 	})
// }

func (q *EventQueue[T]) Pop() *T {
	event := q.TimeQueue.Pop()
	return event
}

func (q *EventQueue[T]) PopAll() []*T {
	eventObj := q.TimeQueue.PopAll()
	result := []*T{}
	for _, i := range eventObj {
		result = append(result, i)
	}
	return result
}

func (q *EventQueue[T]) PopOlderOrEqualThan(t time.Time) []*T {
	eventObj := q.TimeQueue.PopOlderOrEqualThan(t)
	result := []*T{}
	for _, i := range eventObj {
		result = append(result, i)
	}
	return result
}
