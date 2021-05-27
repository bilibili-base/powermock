VERSION=$(shell cat "./VERSION" 2> /dev/null)
GIT_REVISION=$(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_DATE=$$(date +%Y-%m-%d-%H:%M)
GO_FLAGS := -ldflags "-X main.BuildDate=$(BUILD_DATE) -X main.Branch=$(GIT_BRANCH) -X main.Revision=$(GIT_REVISION) -X main.Version=$(VERSION) -extldflags \"-static\" -s -w" -tags netgo
PROTO_FILES=$(shell find . -name *.proto)

# Generate binaries for a powermock release
.PHONY: build
build:
	make build_linux
	make build_windows
	make build_darwin

.PHONY: build_linux
build_linux:
	rm -fr ./dist/powermock-linux-amd64
	mkdir -p ./dist
	GOOS="linux"  GOARCH="amd64" CGO_ENABLED=1 go build $(GO_FLAGS) -o ./dist/powermock-linux-amd64   ./cmd/powermock
	shasum -a 256 ./dist/powermock-linux-amd64  | cut -d ' ' -f 1 > ./dist/powermock-linux-amd64-sha-256

.PHONY: build_windows
build_windows:
	rm -fr ./dist/powermock-windows-amd64
	mkdir -p ./dist
	GOOS="windows"  GOARCH="amd64" CGO_ENABLED=1 go build $(GO_FLAGS) -o ./dist/powermock-windows-amd64   ./cmd/powermock
	shasum -a 256 ./dist/powermock-windows-amd64  | cut -d ' ' -f 1 > ./dist/powermock-windows-amd64-sha-256

.PHONY: build_darwin
build_darwin:
	rm -fr ./dist/powermock-darwin-amd64
	mkdir -p ./dist
	GOOS="darwin"  GOARCH="amd64" CGO_ENABLED=1 go build $(GO_FLAGS) -o ./dist/powermock-darwin-amd64   ./cmd/powermock
	shasum -a 256 ./dist/powermock-darwin-amd64  | cut -d ' ' -f 1 > ./dist/powermock-darwin-amd64-sha-256

.PHONY: proto
proto:
	$(call build_proto_files, $(PROTO_FILES))

lint:
	misspell -error docs
	# Configured via .golangci.yml.
	golangci-lint run

define build_proto_files
@for file in $(1); do \
( 	echo "---\nbuilding: $$file" && \
 	protoc --proto_path=. \
  		--proto_path=$(shell dirname $(shell pwd)) \
  		--proto_path=$(GOPATH)/src \
  		--proto_path=$(GOBIN) \
  		--grpc-gateway_out=. \
  		--go_out=paths=source_relative:. \
  		--go-grpc_out=paths=source_relative:. \
  		--go-errors_out=paths=source_relative:. $$file)  \
done;
endef