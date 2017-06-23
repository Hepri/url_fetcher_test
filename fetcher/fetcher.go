package fetcher

type UrlStatFetcher func(url string) (int, error)
