package fetcher

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testSubstringCount(t *testing.T, url string, substr string, expectedCount int, expectErr bool) {
	fetch := NewUrlSubstringCountFetcher(substr)
	cnt, err := fetch(url)

	if cnt != expectedCount {
		t.Errorf("Substring %s from %s, expected count %d but received %d", substr, url, expectedCount, cnt)
	}

	if expectErr && err == nil {
		t.Errorf("Substring %s from %s, expected error but received nil", substr, url)
	} else if !expectErr && err != nil {
		t.Errorf("Substring %s from %s, error not expected but received %v", substr, url, err)
	}
}

func TestUrlSubstringCountFetcher(t *testing.T) {
	// create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Here are some words: test test TEST Test")
	}))
	defer ts.Close()

	testSubstringCount(t, ts.URL, "test", 2, false)
	testSubstringCount(t, ts.URL, "TEST", 1, false)
	testSubstringCount(t, ts.URL, "Test", 1, false)
	testSubstringCount(t, ts.URL, "unknown", 0, false)
	testSubstringCount(t, ts.URL, "test string with spaces", 0, false)
}

func TestUrlSubstringCountFetcherErrors(t *testing.T) {
	// create unstarted server
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()

	// should receive error
	testSubstringCount(t, ts.URL, "test", 0, true)

	// start server
	ts.Start()

	// shouldn't receive error
	testSubstringCount(t, ts.URL, "test", 0, false)

	// close server
	ts.Close()

	// should receive error
	testSubstringCount(t, ts.URL, "test", 0, true)
}

func BenchmarkSubstringCountFetcher(b *testing.B) {
	// init
	b.StopTimer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()
	fetch := NewUrlSubstringCountFetcher("test")
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		fetch(ts.URL)
	}
}
