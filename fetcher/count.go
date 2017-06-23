package fetcher

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func urlSubstringCountFetcher(url string, substr string) (int, error) {
	res, err := http.Get(url)
	if err != nil || res.Body == nil {
		return 0, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	return strings.Count(string(body), substr), nil
}

func NewUrlSubstringCountFetcher(substr string) UrlStatFetcher {
	return func(url string) (int, error) {
		return urlSubstringCountFetcher(url, substr)
	}
}
