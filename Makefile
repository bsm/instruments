default: test

test:
	go test ./...

test-race:
	go test ./... -race

bench:
	go test ./... -run=NONE -bench=. -benchmem

bench-race:
	go test ./... -run=NONE -bench=. -race

lint:
	golangci-lint run
