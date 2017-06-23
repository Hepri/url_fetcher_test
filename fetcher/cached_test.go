package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
