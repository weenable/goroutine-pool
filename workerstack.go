package goroutine_pool

import "time"

type workerStack struct {
	items []*goWorker
	expiry []*goWorker
	size int32
}

func newWorkerStack(size int32) *workerStack {
	return &workerStack{
		items: make([]*goWorker, 0, size),
		size: size,
	}
}

func (wq *workerStack) insert(worker *goWorker) {
	wq.items = append(wq.items, worker)
}

func (wq *workerStack) detach() *goWorker {
	l := len(wq.items)
	if l == 0 {
		return nil
	}
	
	w := wq.items[l-1]
	wq.items[l-1] = nil //防止内存泄漏
	wq.items = wq.items[:l-1]
	return w
}

func (wq *workerStack) reset() {
	for i := 0; i < len(wq.items); i++ {
		wq.items[i].task <- nil
		wq.items[i] = nil
	}
	wq.items = wq.items[:0]
}

func (wq *workerStack) retrieveExpiry(duration time.Duration) []*goWorker {
	n := len(wq.items)
	if n == 0 {
		return nil
	}
	
	expiryTime := time.Now().Add(-duration)
	index := wq.binarySearch(0, n-1, expiryTime)
	wq.expiry = wq.expiry[:0]
	if index != -1 {
		wq.expiry = append(wq.expiry, wq.items[:index+1]...)
		m := copy(wq.items, wq.items[index+1:])
		for i := m; i < n; i++ {
			wq.items[i] = nil
		}
		wq.items = wq.items[:m]
	}
	return wq.expiry
}

func (wq *workerStack) binarySearch(l, r int, expiryTime time.Time) int {
	var mid int
	for l <= r {
		mid = (l+r)/2
		if expiryTime.Before(wq.items[mid].recycleTime) {
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	return r
}