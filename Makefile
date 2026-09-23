GOHOSTOS := $(shell go env GOHOSTOS)
GOPATH := $(shell go env GOPATH)
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR ?= bin
BUILD_ENV = GOOS=$(GOOS) GOARCH=$(GOARCH)
LDFLAGS ?= -X main.Version=$(VERSION)

PROTOC_GEN_GO_VERSION := v1.34.2
PROTOC_GEN_GO_GRPC_VERSION := v1.5.1
KRATOS_VERSION := v2.8.0
GNOSTIC_VERSION := v0.7.0
WIRE_VERSION := v0.6.0

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES := $(shell $(Git_Bash) -c "find internal -name '*.proto' -print | sort")
	API_PROTO_FILES := $(shell $(Git_Bash) -c "find api -name '*.proto' -print | sort")
else
	INTERNAL_PROTO_FILES := $(shell find internal -name '*.proto' -print | sort)
	API_PROTO_FILES := $(shell find api -name '*.proto' -print | sort)
endif

.PHONY: init config api wire generate tidy build build-win build-linux build-mac test-layout test all help
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	go install github.com/go-kratos/kratos/cmd/kratos/v2@$(KRATOS_VERSION)
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@$(KRATOS_VERSION)
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@$(GNOSTIC_VERSION)
	go install github.com/google/wire/cmd/wire@$(WIRE_VERSION)

# generate internal proto
config:
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./internal \
	       $(INTERNAL_PROTO_FILES)

# generate api proto
api:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --go-http_out=paths=source_relative:./api \
 	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:. \
	       $(API_PROTO_FILES)

# build
build:
	mkdir -p $(BUILD_DIR) && $(BUILD_ENV) go build -ldflags "$(LDFLAGS)" -o ./$(BUILD_DIR)/ ./...

# build Windows amd64
build-win: GOOS := windows
build-win: GOARCH := amd64
build-win: BUILD_DIR := bin/windows
build-win: BUILD_ENV += CGO_ENABLED=0
build-win: LDFLAGS := -s -w -X main.Version=$(VERSION)
build-win: build

# build Linux amd64
build-linux: GOOS := linux
build-linux: GOARCH := amd64
build-linux: BUILD_DIR := bin/linux
build-linux: build

# build macOS arm64
build-mac: GOOS := darwin
build-mac: GOARCH := arm64
build-mac: BUILD_DIR := bin/macos
build-mac: build

# generate Wire dependency injection code
wire: config
	go generate ./...

# generate all protobuf, OpenAPI, and Wire code
generate: api wire

# tidy module dependencies
tidy:
	go mod tidy

# verify that test files are kept under test/
test-layout:
	@files="$$(find . -path './test' -prune -o -name '*_test.go' -print)"; \
	if [ -n "$$files" ]; then \
		echo 'test files must be placed under test/:'; \
		echo "$$files"; \
		exit 1; \
	fi

# test all packages
test: test-layout
	go test -mod=readonly ./...

# generate, build, and test
all: generate
	$(MAKE) build
	$(MAKE) test

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
