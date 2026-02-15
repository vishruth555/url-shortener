.PHONY: run build test fmt

run:
	go run ./cmd/server

build:
	go build ./...

test:
	go test ./...

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './.cache/*')
