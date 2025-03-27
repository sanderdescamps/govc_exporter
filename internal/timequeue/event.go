package timequeue

import "time"

type Event interface {
	GetTimestamp() time.Time
}
