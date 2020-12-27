package goroutine_pool

import (
	"sync"
	"testing"
	"time"
)

const (
	RunTimes           = 10000000
	BenchParam         = 10
	BenchAntsSize      = 200000
	DefaultExpiredTime = 10 * time.Second
)

func demoFunc() {
	time.Sleep(time.Duration(BenchParam) * time.Millisecond)
}

func BenchmarkGoroutines(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			go func() {
				demoFunc()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkPool(b *testing.B) {
	var wg sync.WaitGroup
	p := NewPool(BenchAntsSize)
	defer p.Release()
	
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			_ = p.Submit(func() {
				demoFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
	b.StopTimer()
}