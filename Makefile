test:
	go test ./fetcher ./pool

bench:
	go test -benchmem -bench=. ./fetcher ./pool

