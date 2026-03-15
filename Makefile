VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/natikgadzhi/fm/cmd.Version=$(VERSION)"

.PHONY: build test vet clean

build:
	go build $(LDFLAGS) -o fm .

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f fm
