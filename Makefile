VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/natikgadzhi/fm/cmd.Version=$(VERSION)"

.PHONY: build test vet clean e2e

build:
	go build $(LDFLAGS) -o fm .

test:
	go test ./...

vet:
	go vet ./...

e2e:
	go test -tags e2e -v -timeout 120s ./tests/

clean:
	rm -f fm
