# Justfile for logz project

# Build tasks
build:
	go build -trimpath -ldflags="-s -w -X 'github.com/hermanu/logz/internal/logz.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)' -X 'github.com/hermanu/logz/internal/logz.commit=$(git rev-parse --short HEAD 2>/dev/null || echo none)' -X 'github.com/hermanu/logz/internal/logz.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" -o logz .

install:
	go install -trimpath -ldflags="-s -w -X 'github.com/hermanu/logz/internal/logz.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)' -X 'github.com/hermanu/logz/internal/logz.commit=$(git rev-parse --short HEAD 2>/dev/null || echo none)' -X 'github.com/hermanu/logz/internal/logz.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" .

snapshot:
	bin/goreleaser build --snapshot --clean --single-target

# Test and Lint tasks
test:
	go test -race -count=1 ./...

cover:
	go test -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	echo "Coverage report: coverage.html"

vet:
	go vet ./...

lint:
	bin/golangci-lint run ./...

fmt:
	bin/gofumpt -l -w .

check: vet lint test
	echo "All checks passed!"

# Tooling tasks
tools:
	GOBIN=$(pwd)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2
	GOBIN=$(pwd)/bin go install mvdan.cc/gofumpt@v0.7.0
	GOBIN=$(pwd)/bin go install github.com/goreleaser/goreleaser/v2@v2.5.1

tidy:
	go mod tidy

# Cleanup tasks
clean:
	rm -rf logz dist coverage.out coverage.html bin

# Help
help:
	echo "Build tasks:"
	echo "  build    - Build logz binary"
	echo "  install  - Install logz binary"
	echo "  snapshot - Build snapshot release locally"
	echo ""
	echo "Test and Lint tasks:"
	echo "  test     - Run tests with race detector"
	echo "  cover    - Generate coverage report"
	echo "  vet      - Run go vet"
	echo "  lint     - Run golangci-lint"
	echo "  fmt      - Format code with gofumpt"
	echo "  check    - Run vet, lint, and test"
	echo ""
	echo "Tooling tasks:"
	echo "  tools    - Install development tools"
	echo "  tidy     - Run go mod tidy"
	echo ""
	echo "Cleanup tasks:"
	echo "  clean    - Remove build artifacts"
	echo ""
	echo "Other:"
	echo "  help     - Show this help"