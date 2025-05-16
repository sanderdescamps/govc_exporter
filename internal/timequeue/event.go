package timequeue

import "time"

type Event interface {
	Timestamp() time.Time
}
