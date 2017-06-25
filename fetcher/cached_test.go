package fetcher

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
)

func assertFetchResult(t *testing.T, fn UrlStatFetcher, url string, stat int, err error) {
	a, b := fn(url)
	if a != stat {
		t.Errorf("Expected stat to be %d but received %d", stat, a)
	}

	if b != err {
		t.Errorf("Expected error to be %v but received %v", err, b)
	}
}

func TestConcurrentCachedUrlStatFetcher(t *testing.T) {
	var url1 string = "http://google.ru"
	var url2 string = "http://mail.ru"
	var res1, res2 int = 1, 2

	fetch := NewConcurrentCachedUrlStatFetcher(func(url string) (int, error) {
		if url == url1 {
			return res1, nil
		} else if url == url2 {
			return res2, nil
		} else {
			return -1, nil
		}
	})

	// check cache for different urls
	assertFetchResult(t, fetch, url1, res1, nil)
	assertFetchResult(t, fetch, url2, res2, nil)
	assertFetchResult(t, fetch, url1, res1, nil)
	assertFetchResult(t, fetch, url2, res2, nil)
}

func TestConcurrentCachedUrlStatFetcherConcurrentCall(t *testing.T) {
	var mx sync.Mutex
	var counterMap = make(map[string]int)
	var errUnknown = errors.New("Unknown")

	wrapped := func(url string) (int, error) {
		// first call should return 0, nil
		// second call should return 999, errUnknown
		mx.Lock()
		defer mx.Unlock()

		if counter, ok := counterMap[url]; !ok {
			counterMap[url] = 0

			return 0, nil
		} else {
			counter++
			counterMap[url] = counter

			return counter, errUnknown
		}
	}

	fetch := NewConcurrentCachedUrlStatFetcher(wrapped)

	// run for different urls to be sure
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		// run fetch few times concurrently, they should receive same result (0, nil)
		for j := 0; j < 100; j++ {
			wg.Add(1)
			go func() {
				assertFetchResult(t, fetch, strconv.Itoa(i), 0, nil)
				wg.Done()
			}()
		}
		wg.Wait()
	}

	// clear wrapped func counter map
	counterMap = make(map[string]int)

	// make sure two direct calls of wrapped func will return different results
	var testUrl string = "http://google.ru"
	var res1, res2 int
	var err1, err2 error

	wg.Add(1)
	go func() {
		res1, err1 = wrapped(testUrl)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		res2, err2 = wrapped(testUrl)
		wg.Done()
	}()

	wg.Wait()

	if res1 == res2 {
		t.Errorf("res1 and res2 expected to be different (res1: %d, res2: %d)", res1, res2)
	}
	if err1 == err2 {
		t.Errorf("res1 and res2 expected to be different (err1: %v, err2: %v)", err1, err2)
	}
}

func TestConcurrentCachedUrlStatFetcherSameResult(t *testing.T) {
	// make sure cache save error and result
	var res int = 99123
	var err = errors.New("Some error")

	fetch := NewConcurrentCachedUrlStatFetcher(func(url string) (int, error) {
		return res, err
	})

	res1, err1 := fetch("https://google.com")

	if res1 != res {
		t.Errorf("res1 expected to be %d", res)
	}
	if err1 != err {
		t.Errorf("err1 expected to be %v", err)
	}
}

func BenchmarkConcurrentCachedFetcher(b *testing.B) {
	// init
	b.StopTimer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()
	fetch := NewUrlSubstringCountFetcher("test")
	fetch = NewConcurrentCachedUrlStatFetcher(fetch)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		fetch(ts.URL)
	}
}
