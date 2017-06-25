package fetcher

import "sync"

type urlStateCacheRec struct {
	Stat int
	Err  error
}

type cachedConcurrentFetcher struct {
	sync.Mutex
	fn      UrlStatFetcher
	cache   map[string]urlStateCacheRec
	waiters map[string]chan struct{}
}

func (c *cachedConcurrentFetcher) getCachedValue(url string) (urlStateCacheRec, bool) {
	rec, ok := c.cache[url]
	return rec, ok
}

func (c *cachedConcurrentFetcher) getCachedValueWithLock(url string) (urlStateCacheRec, bool) {
	c.Lock()
	defer c.Unlock()

	return c.getCachedValue(url)
}

func (c *cachedConcurrentFetcher) getOrCreateWaiter(url string) (waiter chan struct{}, found bool) {
	var w chan struct{}
	var ok bool

	if w, ok = c.waiters[url]; !ok {
		w = make(chan struct{})
		c.waiters[url] = w
	}
	return w, ok
}

func (c *cachedConcurrentFetcher) fetch(url string) (int, error) {
	var (
		rec         urlStateCacheRec
		ok          bool
		waiter      chan struct{}
		foundWaiter bool
	)

	// complex lock with two different places to unlock
	c.Lock()
	rec, ok = c.getCachedValue(url)
	if ok {
		// found cached value, lock not needed anymore
		c.Unlock()
		return rec.Stat, rec.Err
	}

	// don't have cached value
	// let's check whether someone is fetching this URL right now or not
	// should be done inside lock
	waiter, foundWaiter = c.getOrCreateWaiter(url)
	c.Unlock()

	if foundWaiter {
		// wait until other goroutine will fetch URL and close waiter
		<-waiter

		rec, _ = c.getCachedValueWithLock(url)
		return rec.Stat, rec.Err
	} else {
		// this is fetching goroutine, call fetch, put value in cache and close waiter
		stat, err := c.fn(url)

		c.Lock()
		c.cache[url] = urlStateCacheRec{
			Stat: stat,
			Err:  err,
		}
		delete(c.waiters, url)
		close(waiter)
		c.Unlock()

		return stat, err
	}
}

// idea of concurrent cached fetcher is to cache URL fetch responses
// and forbid fetch calls for same URL at the same time
// so fetcher will be called only once for each URL
func NewConcurrentCachedUrlStatFetcher(fn UrlStatFetcher) UrlStatFetcher {
	c := &cachedConcurrentFetcher{
		fn:      fn,
		cache:   make(map[string]urlStateCacheRec),
		waiters: make(map[string]chan struct{}),
	}

	return c.fetch
}
