.DEFAULT_GOAL := test

.PHONY: build test test-race bench fmt vet check dump run-sample eval clean

build:
	go install ./cmd/vitals

test:
	go test ./... -count=1

test-race:
	go test -race ./... -count=1

bench:
	go test -bench=. -benchmem ./internal/... -count=1

fmt:
	go fmt ./...

vet:
	go vet ./...

check: fmt vet test

dump: build
	vitals --dump-current

run-sample:
	cat testdata/sample-stdin.json | go run ./cmd/vitals

eval:
	go test ./internal/eval/ -run TestDesignEval -v -count=1

clean:
	rm -rf bin/
	rm -f $$(go env GOPATH)/bin/vitals
