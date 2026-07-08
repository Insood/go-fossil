.DEFAULT_GOAL := build

.PHONY: build run test

build:
	mkdir -p bin .gocache-local
	GOCACHE=$${GOCACHE:-$(CURDIR)/.gocache-local} go build -o bin/go-fossil ./cmd/game

run: build
	./bin/go-fossil

test:
	GOCACHE=$${GOCACHE:-$(CURDIR)/.gocache-local} go test ./...

%:
	@:
