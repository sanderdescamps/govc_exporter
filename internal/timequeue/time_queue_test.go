package timequeue_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	timequeue "github.com/sanderdescamps/govc_exporter/internal/timequeue"
)

type TestItem struct {
	Name      string
	timestamp time.Time
}

func NewTestItem(name string) *TestItem {
	return &TestItem{Name: name, timestamp: time.Now()}
}

func (t TestItem) Timestamp() time.Time {
	return time.Now()
}

func TestTimeQueue_Pop(t *testing.T) {
	q := timequeue.TimeQueue[TestItem]{}
	items := []*TestItem{
		NewTestItem("first"),
		NewTestItem("second"),
		NewTestItem("third"),
		NewTestItem("fourth"),
	}
	q.Add(items[0])
	q.Add(items[2])
	q.Add(items[1])
	q.Add(items[3])

	for {
		item := q.Pop()
		if item != nil {
			fmt.Printf("%s -> %s\n", item.Timestamp().String(), item.Name)
			continue
		}
		break

	}
}

func TestTimeQueue_PopOlderOrEqualThan(t *testing.T) {
	q := timequeue.TimeQueue[TestItem]{}
	items := []*TestItem{
		NewTestItem("first"),
		NewTestItem("second"),
		NewTestItem("third"),
		NewTestItem("fourth"),
	}
	q.Add(items[0])
	q.Add(items[2])
	q.Add(items[1])
	q.Add(items[3])

	now := time.Now()
	older := q.PopOlderOrEqualThan(now.Add(6 * time.Second))
	if len(older) != 2 {
		t.Fatalf("Incorrect number of items popped from queue")
	}

	if q.Len() != 2 {
		t.Fatalf("Incorrect number of items remaining in queue")
	}
	for _, obj := range older {
		t.Logf("POP %s -> %s\n", obj.Timestamp().String(), obj.Name)
	}

	for {
		item := q.Pop()
		if item != nil {
			t.Logf("REMAIN %s -> %s\n", item.Timestamp().String(), item.Name)
			continue
		}
		break
	}
}

func TestTimeQueue_PopAllItems(t *testing.T) {
	q := timequeue.TimeQueue[TestItem]{}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(5 * time.Microsecond)
			item := fmt.Sprintf("worker 1.%d", i)
			q.Add(NewTestItem(item))
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(5 * time.Microsecond)
			item := fmt.Sprintf("worker 2.%d", i)
			q.Add(NewTestItem(item))
		}
		wg.Done()
	}()

	time.Sleep(100 * time.Microsecond)
	for i := range q.PopAllItems() {
		fmt.Printf("%s\n", i.Name)
	}

	wg.Wait()
	fmt.Print(strings.Repeat("-", 50) + "\n")
	for i := range q.PopAllItems() {
		fmt.Printf("%s\n", i.Name)
	}
}
