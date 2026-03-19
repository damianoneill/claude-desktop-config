BIN     := claude-desktop-config
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/damianoneill/claude-desktop-config/cmd.version=$(VERSION)"

.DEFAULT_GOAL := help

.PHONY: help
help: ## show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sed 's/:.*## /|/' | awk -F'|' '{printf "  %-20s %s\n", $$1, $$2}'

.PHONY: build
build: ## build the binary
	go build $(LDFLAGS) -o $(BIN) .

.PHONY: test
test: ## run all tests
	go test ./...

.PHONY: lint
lint: ## run golangci-lint
	golangci-lint run

.PHONY: fmt
fmt: ## format source code
	gofmt -w .

.PHONY: clean
clean: ## remove build artifacts
	rm -f $(BIN)

.PHONY: install
install: ## install binary to GOPATH/bin
	go install $(LDFLAGS) .

.PHONY: release-dry
release-dry: ## local goreleaser snapshot (no publish)
	goreleaser release --snapshot --clean
