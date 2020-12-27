package goroutine_pool

import (
	"runtime"
	"time"
)

type goWorker struct {
	pool *Pool
	task chan func()
	recycleTime time.Time
}

func (w *goWorker) run() {
	w.pool.incRunning()
	go func() {
		// 清理回收
		defer func() {
			w.pool.decRunning()
			//放回到sync.Pool中
			w.pool.workerCache.Put(w)
			if p := recover(); p != nil {
				w.pool.logger.Printf("worker 异常退出：%v\n", p)
				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				w.pool.logger.Printf("当前栈信息：%s\n", string(buf[:n]))
			}
		}()
		
		for f := range w.task {
			if f == nil {
				//收到退出信消息
				return
			}
			
			//执行任务
			f()
			
			//归还worker到活跃goroutine队列
			if ok := w.pool.revertWorker(w); !ok {
				//没有归还到goroutine活跃队列中，直接销毁
				return
			}
		}
	}()
}