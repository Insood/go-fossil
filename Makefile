.DEFAULT_GOAL := build

.PHONY: build build-windows run test

build:
	mkdir -p bin .gocache-local
	GOCACHE=$${GOCACHE:-$(CURDIR)/.gocache-local} go build -o bin/go-fossil ./cmd/game

build-windows:
	mkdir -p bin/win .gocache-local
	GOOS=windows GOARCH=amd64 GOCACHE=$${GOCACHE:-$(CURDIR)/.gocache-local} go build -o bin/win/go-fossil.exe ./cmd/game

run: build
	./bin/go-fossil

test:
	GOCACHE=$${GOCACHE:-$(CURDIR)/.gocache-local} go test ./...

%:
	@:
