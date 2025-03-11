package pool_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/sanderdescamps/govc_exporter/collector/pool"
)

type TestPoolItem struct {
	Value string
}

func TestMultiAccessPool(t *testing.T) {
	poolSize := 4
	iterations := 10

	p := pool.NewMultiAccessPool[TestPoolItem](&TestPoolItem{Value: "item 1"}, poolSize, nil)

	releaseFunc := make(chan func(), poolSize)
	go func() {
		for i := 0; i < iterations; i++ {
			_, release, err := p.Acquire()
			if err != nil {
				t.Fatalf(err.Error())
			}
			fmt.Printf("acquire item=%d\n", i)
			releaseFunc <- release
		}
		fmt.Printf("shutdown starter\n")
	}()

	time.Sleep(10 * time.Microsecond)
	fmt.Print("start releasing\n")

	for i := 0; i < iterations; i++ {
		release := <-releaseFunc
		release()
		fmt.Printf("release item=%d\n", i)
		time.Sleep(1 * time.Second)
	}
}
