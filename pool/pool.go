package pool

import (
	"context"
	"sync"
)

type WorkerFn func(string)

type UrlWorkerPool struct {
	ctx       context.Context
	urls      chan string
	quotaChan chan struct{}
	fn        WorkerFn

	running sync.WaitGroup
}

func (w *UrlWorkerPool) Wait() {
	w.running.Wait()
}

func (w *UrlWorkerPool) acquireQuota() {
	w.running.Add(1)
	<-w.quotaChan
}

func (w *UrlWorkerPool) releaseQuota() {
	w.quotaChan <- struct{}{}
	w.running.Done()
}

func (w *UrlWorkerPool) receiveUrls() {
	defer w.running.Done()
	for {
		select {
		case <-w.ctx.Done():
			return
		case u, ok := <-w.urls:
			if !ok {
				// closed channel
				return
			} else {
				w.acquireQuota()

				go func(url string) {
					w.fn(url)
					w.releaseQuota()
				}(u)
			}
		}
	}
}

func NewUrlWorkerPool(ctx context.Context, urls chan string, quota int, fn WorkerFn) *UrlWorkerPool {
	wp := &UrlWorkerPool{
		ctx:       ctx,
		urls:      urls,
		quotaChan: make(chan struct{}, quota),
		fn:        fn,
	}

	// init quota chan
	for i := 0; i < quota; i++ {
		wp.quotaChan <- struct{}{}
	}

	wp.running.Add(1)
	go wp.receiveUrls()

	return wp
}
