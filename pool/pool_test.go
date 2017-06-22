package pool

import (
	"context"
	"sync"
	"testing"
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
