package timequeue_test

import (
	"fmt"
	"testing"
	"time"

	timequeue "github.com/sanderdescamps/govc_exporter/internal/timequeue"
)

func TestTimeQueue_Pop(t *testing.T) {
	q := timequeue.TimeQueue[string]{}
	items := []string{"first", "second", "third", "fourth"}
	q.Add(time.Now().Add(5*time.Second), &items[0])
	q.Add(time.Now().Add(7*time.Second), &items[2])
	q.Add(time.Now().Add(6*time.Second), &items[1])
	q.Add(time.Now().Add(8*time.Second), &items[3])

	for {
		timestamp, item := q.Pop()
		if item != nil {
			fmt.Printf("%s -> %s\n", timestamp.String(), *item)
			continue
		}
		break

	}
}

func TestTimeQueue_PopOlderOrEqualThan(t *testing.T) {
	q := timequeue.TimeQueue[string]{}
	items := []string{"first", "second", "third", "fourth"}
	now := time.Now()
	q.Add(now.Add(5*time.Second), &items[0])
	q.Add(now.Add(7*time.Second), &items[2])
	q.Add(now.Add(6*time.Second), &items[1])
	q.Add(now.Add(8*time.Second), &items[3])

	older := q.PopOlderOrEqualThan(now.Add(6 * time.Second))
	if len(older) != 2 {
		t.Fatalf("Incorrect number of items popped from queue")
	}

	if q.Len() != 2 {
		t.Fatalf("Incorrect number of items remaining in queue")
	}
	for _, obj := range older {
		t.Logf("POP %s -> %s\n", obj.Timestamp.String(), *obj.Obj)
	}

	for {
		timestamp, item := q.Pop()
		if item != nil {
			t.Logf("REMAIN %s -> %s\n", timestamp.String(), *item)
			continue
		}
		break
	}
}
