SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

GO          ?= go
BIN_DIR     := $(CURDIR)/bin
GOLANGCI    := $(BIN_DIR)/golangci-lint
GOFUMPT     := $(BIN_DIR)/gofumpt
GORELEASER  := $(BIN_DIR)/goreleaser

LDFLAGS := -s -w \
	-X 'github.com/hermanu/logz/internal/logz.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)' \
	-X 'github.com/hermanu/logz/internal/logz.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo none)' \
	-X 'github.com/hermanu/logz/internal/logz.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)'

##@ General

.PHONY: help
help: ## Show this help.
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} \
	/^[a-zA-Z0-9_.-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
	/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Build

.PHONY: build
build: ## Build the logz binary into ./logz.
	$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o logz .

.PHONY: install
install: ## Install logz into $$GOBIN / $$GOPATH/bin.
	$(GO) install -trimpath -ldflags="$(LDFLAGS)" .

.PHONY: snapshot
snapshot: $(GORELEASER) ## Build a snapshot release locally (no publish).
	$(GORELEASER) build --snapshot --clean

##@ Test & Lint

.PHONY: test
test: ## Run unit tests with race detector.
	$(GO) test -race -count=1 ./...

.PHONY: cover
cover: ## Run tests and produce coverage.html.
	$(GO) test -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: vet
vet: ## Run go vet.
	$(GO) vet ./...

.PHONY: lint
lint: $(GOLANGCI) ## Run golangci-lint.
	$(GOLANGCI) run ./...

.PHONY: fmt
fmt: $(GOFUMPT) ## Format code with gofumpt.
	$(GOFUMPT) -l -w .

.PHONY: check
check: vet lint test ## Run vet + lint + test.

##@ Tooling

.PHONY: tools
tools: $(GOLANGCI) $(GOFUMPT) $(GORELEASER) ## Install dev tools into ./bin.

$(GOLANGCI):
	GOBIN=$(BIN_DIR) $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2

$(GOFUMPT):
	GOBIN=$(BIN_DIR) $(GO) install mvdan.cc/gofumpt@v0.7.0

$(GORELEASER):
	GOBIN=$(BIN_DIR) $(GO) install github.com/goreleaser/goreleaser/v2@v2.5.1

##@ Housekeeping

.PHONY: tidy
tidy: ## go mod tidy.
	$(GO) mod tidy

.PHONY: clean
clean: ## Remove build artifacts.
	rm -rf logz dist coverage.out coverage.html bin
