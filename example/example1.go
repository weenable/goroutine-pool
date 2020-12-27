package main

import (
	"fmt"
	goroutine_pool "github.com/weenable/goroutine-pool"
	"sync"
	"time"
)
func demoFunc() {
	time.Sleep(10 * time.Millisecond)
	fmt.Println("Hello World!")
}

func main() {
	defer goroutine_pool.Release()
	var wg sync.WaitGroup
	
	syncCalculateSum := func() {
		demoFunc()
		wg.Done()
	}
	
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		_ = goroutine_pool.Submit(syncCalculateSum)
	}
	wg.Wait()
	fmt.Printf("running goroutines: %d\n", goroutine_pool.Running())
	fmt.Printf("finish all tasks.\n")
}