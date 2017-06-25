package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type initPoolFn func(chan string, int, WorkerFn) chan struct{}

func initBasicPool(urls chan string, quota int, fn WorkerFn) chan struct{} {
	wg := sync.WaitGroup{}
	for i := 0; i < quota; i++ {
		wg.Add(1)
		go func() {
			for url := range urls {
				fn(url)
			}
			wg.Done()
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	return done
}

func initUrlWorkerPool(urls chan string, quota int, fn WorkerFn) chan struct{} {
	p := NewUrlWorkerPool(context.Background(), urls, quota, fn)

	done := make(chan struct{})
	go func() {
		p.Wait()
		close(done)
	}()

	return done
}

func benchPoolWithQuota(b *testing.B, quota int, initPool initPoolFn) {
	b.StopTimer()

	urls := make(chan string)
	done := initPool(urls, quota, func(string) {})

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		urls <- "test"
	}

	close(urls)
	<-done
}

func BenchmarkBasicPool10(b *testing.B) {
	benchPoolWithQuota(b, 10, initBasicPool)
}

func BenchmarkBasicPool100(b *testing.B) {
	benchPoolWithQuota(b, 100, initBasicPool)
}

func BenchmarkBasicPool1000(b *testing.B) {
	benchPoolWithQuota(b, 1000, initBasicPool)
}

func BenchmarkUrlWorkerPool10(b *testing.B) {
	benchPoolWithQuota(b, 10, initUrlWorkerPool)
}

func BenchmarkUrlWorkerPool100(b *testing.B) {
	benchPoolWithQuota(b, 100, initUrlWorkerPool)
}

func BenchmarkUrlWorkerPool1000(b *testing.B) {
	benchPoolWithQuota(b, 1000, initUrlWorkerPool)
}

func TestUrlWorkerPool(t *testing.T) {
	const totalCount int = 10000

	var callsCount int64
	var urls = make(chan string)

	w := NewUrlWorkerPool(context.Background(), urls, 2, func(url string) {
		atomic.AddInt64(&callsCount, 1)
	})

	go func() {
		for i := 0; i < totalCount; i++ {
			urls <- "test"
		}
		close(urls)
	}()

	w.Wait()

	if int64(totalCount) != callsCount {
		t.Errorf("Calls count expected to be %d", totalCount)
	}
}

func TestUrlWorkerPoolConcurrency(t *testing.T) {
	const maxConcurrency = 5
	var urls = make(chan string)

	var mx sync.Mutex
	var counter int
	var foundMax int

	w := NewUrlWorkerPool(context.Background(), urls, maxConcurrency, func(url string) {
		mx.Lock()
		counter++
		if counter > foundMax {
			foundMax = counter
		}
		mx.Unlock()

		// some long work
		time.Sleep(time.Millisecond * 100)

		mx.Lock()
		counter--
		mx.Unlock()
	})

	go func() {
		for i := 0; i < 100; i++ {
			urls <- "t"
		}
		close(urls)
	}()

	w.Wait()

	if foundMax > maxConcurrency {
		t.Errorf("Max concurrency quota exceeded")
	} else if foundMax < maxConcurrency {
		t.Errorf("Max concurrency is not reached")
	}
}
