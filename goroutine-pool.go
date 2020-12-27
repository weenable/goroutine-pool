package goroutine_pool

import (
	"errors"
	"log"
	"math"
	"os"
)

var (
	defaultPool = NewPool(math.MaxInt32)
	defaultLogger = log.New(os.Stderr, "", log.LstdFlags)
	
	ErrPoolClosed = errors.New("goroutine协程池已经关闭")
	ErrPoolOverload = errors.New("goroutine协程池已经没有可用对象")
)

func Submit(task func()) error {
	return defaultPool.Submit(task)
}

func Running() int32 {
	return defaultPool.Running()
}

func Release() {
	defaultPool.Release()
}