package main

import (
	"bufio"
	"context"
	"os"

	"fmt"

	"sync/atomic"

	"github.com/Hepri/url_fetcher/fetcher"
	"github.com/Hepri/url_fetcher/pool"
)

const workerQuota = 5

func main() {
	urls := make(chan string)

	var total int64

	fetch := fetcher.NewUrlSubstringCountFetcher("Go")
	fetch = fetcher.NewConcurrentCachedUrlStatFetcher(fetch)

	w := pool.NewUrlWorkerPool(context.Background(), urls, workerQuota, func(url string) {
		cnt, err := fetch(url)
		if err != nil {
			fmt.Printf("Cannot fetch from %s: %s\n", url, err)
			return
		}
		fmt.Printf("%s: %d\n", url, cnt)

		atomic.AddInt64(&total, int64(cnt))
	})

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		urls <- scanner.Text()
	}
	close(urls)

	w.Wait()

	fmt.Printf("Total: %d", total)
}
