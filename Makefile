BIN     := claude-desktop-config
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/damianoneill/claude-desktop-config/cmd.version=$(VERSION)"

.PHONY: build test lint fmt clean install release-dry

build:
	go build $(LDFLAGS) -o $(BIN) .

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .

clean:
	rm -f $(BIN)

install:
	go install $(LDFLAGS) .

release-dry:
	goreleaser release --snapshot --clean
