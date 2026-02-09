.PHONY: build build-all test lint fmt clean install docker docker-buildx-setup docker-login docker-push docker-publish version tag-version release ci

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

# Docker Hub configuration
DOCKER_REPO ?= portertech/skills-mcp-server
DOCKER_PLATFORMS ?= linux/amd64,linux/arm64
DOCKER_BUILDER ?= skills-builder

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

# Setup buildx builder for multi-arch builds
docker-buildx-setup:
	@docker buildx inspect $(DOCKER_BUILDER) >/dev/null 2>&1 || \
		docker buildx create --name $(DOCKER_BUILDER) --driver docker-container --bootstrap
	@docker buildx use $(DOCKER_BUILDER)

# Login to Docker Hub (requires DOCKER_USERNAME and DOCKER_PASSWORD env vars)
docker-login:
	@echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin

# Build and push multi-arch image to Docker Hub
docker-push: docker-buildx-setup
	docker buildx build \
		--platform $(DOCKER_PLATFORMS) \
		--tag $(DOCKER_REPO):$(VERSION) \
		--tag $(DOCKER_REPO):latest \
		--push .

# Publish to Docker Hub (assumes already logged in, or use docker-login first)
docker-publish: docker-push

# List skills in testdata
list-test:
	go run ./cmd/skills --list ./testdata/skills

# Run all checks (used in CI)
ci: lint test

# List version
version:
	@echo $(VERSION)

# Create annotated git tag (requires VERSION to be set explicitly)
tag-version:
	@if [ "$(VERSION)" = "dev" ] || echo "$(VERSION)" | grep -q dirty; then \
		echo "Error: VERSION must be set explicitly and not be 'dev' or dirty"; \
		echo "Usage: make tag-version VERSION=1.0.0"; \
		exit 1; \
	fi
	git tag -a "v$(VERSION)" -m "v$(VERSION)"

# Full release: ci -> tag-version -> docker-publish
# Usage: make release VERSION=1.0.0
release: ci tag-version docker-publish
