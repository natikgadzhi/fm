.PHONY: build test vet lint e2e clean

BINARY_NAME := fm
BUILD_DIR := .

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X github.com/natikgadzhi/fm/cmd.Version=$(VERSION) \
           -X github.com/natikgadzhi/fm/cmd.Commit=$(COMMIT) \
           -X github.com/natikgadzhi/fm/cmd.Date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) .

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

e2e:
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; \
	go test -tags e2e -v -timeout 120s ./tests/

clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -rf dist/
