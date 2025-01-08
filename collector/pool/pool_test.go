package pool_test

type TestPoolItem struct {
	Value string
}

// func TestPool(t *testing.T) {
// 	p := pool.NewFixedSizePoolFromCreatorFunc[TestPoolItem](func() *TestPoolItem {
// 		return &TestPoolItem{Value: uuid.New().String()}

// 	}, 8)

// 	taken := make(chan int, 10)
// 	go func() {
// 		for i := 0; i < 10; i++ {
// 			s, id := p.Acquire()
// 			fmt.Printf("pool item ID=%d item=%s\n", id, s.Value)
// 			taken <- id
// 		}
// 		fmt.Printf("shutdown starter\n")
// 	}()

// 	time.Sleep(10 * time.Microsecond)
// 	fmt.Print("start releasing\n")

// 	go func() {
// 		for id := range taken {
// 			fmt.Printf("release %d\n", id)
// 			p.Release(id)
// 			time.Sleep(time.Second)
// 		}
// 	}()
// 	time.Sleep(5 * time.Second)
// }
