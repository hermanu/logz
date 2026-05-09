# Justfile for logz project - public repository safe

# Build the binary
build:
	go build -trimpath -ldflags="-s -w -X 'github.com/hermanu/logz/internal/logz.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)' -X 'github.com/hermanu/logz/internal/logz.commit=$(git rev-parse --short HEAD 2>/dev/null || echo none)' -X 'github.com/hermanu/logz/internal/logz.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" -o logz .

# Install the binary
install:
	go install -trimpath -ldflags="-s -w -X 'github.com/hermanu/logz/internal/logz.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)' -X 'github.com/hermanu/logz/internal/logz.commit=$(git rev-parse --short HEAD 2>/dev/null || echo none)' -X 'github.com/hermanu/logz/internal/logz.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" .

# Run tests with race detector
test:
	go test -race -count=1 ./...

# Generate coverage report
cover:
	mkdir -p coverage
	go test -race -count=1 -covermode=atomic -coverprofile=./coverage/coverage.txt ./...
	go tool cover -html=./coverage/coverage.txt -o ./coverage/coverage.html
	echo "Coverage report: ./coverage/coverage.html"

# Run go vet
vet:
	go vet ./...

# Run golangci-lint
lint:
	bin/golangci-lint run ./...

# Format code with gofumpt
fmt:
	bin/gofumpt -l -w .

# Run all checks: vet, lint, test
check: vet lint test
	echo "All checks passed!"

# Install development tools
tools:
	GOBIN=$(pwd)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2
	GOBIN=$(pwd)/bin go install mvdan.cc/gofumpt@v0.7.0
	GOBIN=$(pwd)/bin go install github.com/goreleaser/goreleaser/v2@v2.5.1

# Tidy go.mod
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -rf logz dist coverage bin

# Build snapshot release locally
snapshot:
	bin/goreleaser build --snapshot --clean --single-target

# Help
help:
	just --list