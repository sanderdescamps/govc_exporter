package timequeue

import (
	"time"
)

type EventQueue[T interface {
	GetTimestamp() time.Time
}] struct {
	TimeQueue[T]
}

func NewEventQueue[T interface {
	GetTimestamp() time.Time
}]() *EventQueue[T] {
	return &EventQueue[T]{
		TimeQueue: TimeQueue[T]{},
	}
}

func (q *EventQueue[T]) Add(e T) {
	timestamp := e.GetTimestamp()
	q.TimeQueue.Add(timestamp, &e)
}

func (q *EventQueue[T]) Pop() *T {
	_, event := q.TimeQueue.pop()
	return event
}

func (q *EventQueue[T]) PopAll() []*T {
	eventObj := q.TimeQueue.PopAll()
	result := []*T{}
	for _, i := range eventObj {
		result = append(result, i.Obj)
	}
	return result
}

func (q *EventQueue[T]) PopOlderOrEqualThan(t time.Time) []*T {
	eventObj := q.TimeQueue.PopOlderOrEqualThan(t)
	result := []*T{}
	for _, i := range eventObj {
		result = append(result, i.Obj)
	}
	return result
}
