NAME := vistio

GO_ENV ?= CGO_ENABLED=0
GO_PKGS := $(shell go list ./... | grep -vE "(vendor|example)")

BUILD_VERSION ?= $(shell git describe --tags)
BUILD_BRANCH ?= $(shell git rev-parse --abbrev-ref @)
BUILD_TIMESTAMP ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS_PREFIX := -X github.com/nghialv/vistio/version
BUILD_OPTS ?= -ldflags "$(LDFLAGS_PREFIX).Version=$(BUILD_VERSION) $(LDFLAGS_PREFIX).Branch=$(BUILD_BRANCH) $(LDFLAGS_PREFIX).BuildTimestamp=$(BUILD_TIMESTAMP) -w"

.PHONY: build
build: BUILD_DIR ?= ./build
build: BUILD_ENV ?= GOOS=linux GOARCH=amd64
build:
	$(BUILD_ENV) $(GO_ENV) go build $(BUILD_OPTS) -o $(BUILD_DIR)/$(NAME) ./cmd/vistio/main.go

PHONY: test
test:
	$(GO_ENV) go  test $(GO_PKGS)
