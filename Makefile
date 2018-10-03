default: vet test

test:
	go test ./...

test-race:
	go test ./... -race

vet:
	go vet ./...

bench:
	go test ./... -run=NONE -bench=. -benchmem

bench-race:
	go test ./... -run=NONE -bench=. -race
