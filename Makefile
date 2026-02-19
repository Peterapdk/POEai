.PHONY: build test lint install

build:
	go build ./cmd/...

test:
	go test ./... -race -v

lint:
	golangci-lint run ./...

install:
	go build -o ~/.local/bin/poe-gateway ./cmd/poe-gateway/
	go build -o ~/.local/bin/poe ./cmd/poe/
	go build -o ~/.local/bin/poe-node ./cmd/poe-node/
