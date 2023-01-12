GIT_COMMIT = $(shell git rev-list -1 HEAD)
BINARY_WASM = pixlet.wasm
LDFLAGS_WASM = -ldflags="-s -w -X 'tidbyt.dev/pixlet/cmd.Version=$(GIT_COMMIT)'"

ifeq ($(OS),Windows_NT)
	BINARY = pixlet.exe
	LDFLAGS = -ldflags="-s -extldflags=-static -X 'tidbyt.dev/pixlet/cmd.Version=$(GIT_COMMIT)'"
else
	BINARY = pixlet
	LDFLAGS = -ldflags="-X 'tidbyt.dev/pixlet/cmd.Version=$(GIT_COMMIT)'"
endif

all: build

test:
	go test -v -cover ./...

clean:
	rm -f $(BINARY)
	rm -rf ./build
	rm -rf ./out

bench:
	go test -benchmem -benchtime=20s -bench BenchmarkRunAndRender tidbyt.dev/pixlet/encode

build:
	 go build $(LDFLAGS) -o $(BINARY) tidbyt.dev/pixlet

build-wasm:
	GOOS=js GOARCH=wasm go build $(LDFLAGS_WASM) -o $(BINARY_WASM) tidbyt.dev/pixlet/wasm/main

embedfonts:
	go run render/gen/embedfonts.go
	gofmt -s -w ./

widgets:
	 go run runtime/gen/main.go
	 gofmt -s -w ./

release-macos: clean
	./scripts/release-macos.sh

release-linux: clean
	./scripts/release-linux.sh

release-windows: clean
	./scripts/release-windows.sh

install-buildifier:
	go install github.com/bazelbuild/buildtools/buildifier@latest

lint:
	@ buildifier --version >/dev/null 2>&1 || $(MAKE) install-buildifier
	buildifier -r ./

format: lint