.DEFAULT_GOAL := build

.PHONY: build build-windows run test

LINUX_BIN_DIR := bin/linux
WINDOWS_BIN_DIR := bin/win
ASSET_SOURCE_DIR := cmd/game/assets

build:
	mkdir -p bin
	mkdir -p $(LINUX_BIN_DIR) .gocache-local
	find $(LINUX_BIN_DIR) -mindepth 1 -exec rm -rf {} +
	find bin -mindepth 1 -maxdepth 1 ! -name linux ! -name win -exec rm -rf {} +
	cp -R $(ASSET_SOURCE_DIR) $(LINUX_BIN_DIR)/assets
	GOCACHE=$${GOCACHE:-$(CURDIR)/.gocache-local} go build -o $(LINUX_BIN_DIR)/go-fossil ./cmd/game

build-windows:
	mkdir -p bin
	mkdir -p $(WINDOWS_BIN_DIR) .gocache-local
	find $(WINDOWS_BIN_DIR) -mindepth 1 -exec rm -rf {} +
	find bin -mindepth 1 -maxdepth 1 ! -name linux ! -name win -exec rm -rf {} +
	cp -R $(ASSET_SOURCE_DIR) $(WINDOWS_BIN_DIR)/assets
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=$${CC:-x86_64-w64-mingw32-gcc} GOCACHE=$${GOCACHE:-$(CURDIR)/.gocache-local} go build -o $(WINDOWS_BIN_DIR)/go-fossil.exe ./cmd/game

run: build
	./$(LINUX_BIN_DIR)/go-fossil

test:
	GOCACHE=$${GOCACHE:-$(CURDIR)/.gocache-local} go test ./...

%:
	@:
