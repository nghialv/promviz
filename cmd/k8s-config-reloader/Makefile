NAME := k8s-config-reloader

GO_ENV ?= CGO_ENABLED=0
BUILD_VERSION ?= $(shell git describe --tags)
LDFLAGS_PREFIX := -X main.Version=$(BUILD_VERSION)
BUILD_OPTS ?= -ldflags "$(LDFLAGS_PREFIX)"

.PHONY: build
build: BUILD_DIR ?= ./build
build: BUILD_ENV ?= GOOS=linux GOARCH=amd64
build:
	$(BUILD_ENV) $(GO_ENV) go build $(BUILD_OPTS) -o $(BUILD_DIR)/$(NAME) ./main.go
