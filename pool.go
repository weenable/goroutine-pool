package goroutine_pool

import (
	"github.com/weenable/goroutine-pool/internal"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	OPENED = iota
	CLOSED
)

type Pool struct {
	expiryDuration time.Duration
	logger *log.Logger
	lock sync.Locker
	size int32
	workerCache sync.Pool
	state int32
	workers *workerStack
	running int32
}

func (p *Pool) purgePeriodically() {
	heatbeat := time.NewTicker(p.expiryDuration)
	defer heatbeat.Stop()
	
	for range heatbeat.C {
		if atomic.LoadInt32(&p.state) == CLOSED {
			break
		}
		
		p.lock.Lock()
		expiredWorkers := p.workers.retrieveExpiry(p.expiryDuration)
		p.lock.Unlock()
		
		for i := range expiredWorkers {
			expiredWorkers[i].task <- nil
			expiredWorkers[i] = nil
		}
		
	}
}

func NewPool(size int32) (*Pool) {
	p := &Pool{
		expiryDuration: 10 * time.Second, //10秒启动一次清理任务
		logger: defaultLogger,
		lock: internal.NewSpinLock(),
		size: size,
	}
	
	//sync.pool初始化
	p.workerCache.New = func() interface{} {
		return &goWorker{
			pool: p,
			task: make(chan func(), 1),
		}
	}
	
	//活跃worker列表
	p.workers = newWorkerStack(size)
	
	//清理任务
	go p.purgePeriodically()
	return p
}

func (p *Pool) Submit(task func()) error {
	if atomic.LoadInt32(&p.state) == CLOSED {
		return ErrPoolClosed
	}
	
	var w *goWorker
	if w = p.retrieveWorker(); w == nil {
		 return ErrPoolOverload
	}
	w.task <- task
	return nil
}

func (p *Pool) Release() {
	atomic.StoreInt32(&p.state, CLOSED)
	p.lock.Lock()
	defer p.lock.Unlock()
	p.workers.reset()
}

func (p *Pool) Running() int32 {
	return atomic.LoadInt32(&p.running)
}

func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *Pool) retrieveWorker() (w *goWorker) {
	spawnWorker := func() {
		w = p.workerCache.Get().(*goWorker)
		w.run()
	}
	
	p.lock.Lock()
	w = p.workers.detach()
	if w != nil {
		p.lock.Unlock()
	} else if size := p.size; size == -1 {
		// 不使用活跃goroutine队列
		p.lock.Unlock()
		spawnWorker()
	} else if p.Running() < size {
		// 从sync.Pool中拿新的goroutine
		spawnWorker()
	}
	return
}

func (p *Pool) revertWorker(worker *goWorker) bool {
	if (p.size >0 && p.Running() > p.size){
		// 不用还了，直接回收
		return false
	}
	
	worker.recycleTime = time.Now()
	
	//goroutine池已经关闭
	p.lock.Lock()
	defer p.lock.Unlock()
	if atomic.LoadInt32(&p.state) == CLOSED {
		return false
	}
	
	p.workers.insert(worker)
	return true
	
	
}