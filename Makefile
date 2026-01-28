.PHONY: build build-all test lint fmt clean install docker

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

# Build the binary
build:
	go build $(LDFLAGS) -o skills ./cmd/skills

# Build everything (binary + docker)
build-all: build docker

# Run all tests
test:
	go test ./...

# Run tests with verbose output
test-v:
	go test -v ./...

# Run tests with coverage
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	gofmt -w .

# Check formatting (fails if not formatted)
fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files not formatted:"; gofmt -l .; exit 1)

# Run go vet
vet:
	go vet ./...

# All linting checks
lint: fmt-check vet

# Clean build artifacts
clean:
	rm -f skills coverage.out coverage.html

# Install to GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/skills

# Build Docker image
docker:
	docker build -t skills-mcp-server:latest .

# Build Docker image with version tag
docker-release:
	docker build -t skills-mcp-server:$(VERSION) -t skills-mcp-server:latest .

# List skills in testdata
list-test:
	go run ./cmd/skills --list ./testdata/skills

# Run all checks (used in CI)
ci: lint test
