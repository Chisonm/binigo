.PHONY: help build install test bench lint clean dev docs

# Variables
BINARY_NAME=binigo
VERSION=$(shell git describe --tags --always --dirty)
BUILD_DIR=./bin
GO=go
GOFLAGS=-ldflags "-X main.version=$(VERSION)"

# Colors for terminal output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

help: ## Show this help message
	@echo '$(GREEN)Binigo Framework - Available Commands$(NC)'
	@echo ''
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

dev: ## Run in development mode with hot reload
	@echo "$(GREEN)Starting development server...$(NC)"
	air

install-dev-tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(NC)"
	$(GO) install github.com/cosmtrek/air@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install golang.org/x/tools/cmd/goimports@latest

##@ Building

build: ## Build the CLI binary
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/binigo
	@echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

build-all: ## Build for all platforms
	@echo "$(GREEN)Building for all platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/binigo
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/binigo
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/binigo
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/binigo
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/binigo
	@echo "$(GREEN)✓ All builds complete$(NC)"

install: build ## Install the CLI binary to $GOPATH/bin
	@echo "$(GREEN)Installing $(BINARY_NAME)...$(NC)"
	$(GO) install $(GOFLAGS) ./cmd/binigo
	@echo "$(GREEN)✓ Installed successfully$(NC)"

##@ Testing

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	$(GO) test -v -race -timeout 30s ./...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

test-integration: ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(NC)"
	$(GO) test -v -race -tags=integration ./...

bench: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GO) test -bench=. -benchmem -run=^$$ ./...

##@ Code Quality

lint: ## Run linters
	@echo "$(GREEN)Running linters...$(NC)"
	golangci-lint run --timeout=5m

fmt: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GO) fmt ./...
	goimports -w .

vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GO) vet ./...

check: fmt vet lint test ## Run all checks (format, vet, lint, test)

##@ Maintenance

clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	$(GO) clean
	@echo "$(GREEN)✓ Cleaned$(NC)"

deps: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod verify

tidy: ## Tidy dependencies
	@echo "$(GREEN)Tidying dependencies...$(NC)"
	$(GO) mod tidy

update-deps: ## Update dependencies
	@echo "$(GREEN)Updating dependencies...$(NC)"
	$(GO) get -u ./...
	$(GO) mod tidy

##@ Documentation

docs: ## Generate documentation
	@echo "$(GREEN)Generating documentation...$(NC)"
	$(GO) doc -all ./... > docs/API.md
	@echo "$(GREEN)✓ Documentation generated$(NC)"

serve-docs: ## Serve documentation locally
	@echo "$(GREEN)Serving documentation on http://localhost:6060$(NC)"
	godoc -http=:6060

##@ Release

tag: ## Create a new git tag (usage: make tag VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)VERSION is required. Usage: make tag VERSION=v1.0.0$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Creating tag $(VERSION)...$(NC)"
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "$(GREEN)✓ Tag $(VERSION) created and pushed$(NC)"

release: build-all ## Create a release (builds all platforms)
	@echo "$(GREEN)Creating release artifacts...$(NC)"
	cd $(BUILD_DIR) && \
	tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
	tar -czf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
	tar -czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
	tar -czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
	zip $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "$(GREEN)✓ Release artifacts created in $(BUILD_DIR)$(NC)"

##@ Docker

docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t binigo:$(VERSION) -t binigo:latest .

docker-run: ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker run -p 8080:8080 binigo:latest

##@ Miscellaneous

version: ## Show version
	@echo "$(GREEN)Binigo Framework$(NC)"
	@echo "Version: $(VERSION)"
	@echo "Go Version: $(shell $(GO) version)"

stats: ## Show project statistics
	@echo "$(GREEN)Project Statistics:$(NC)"
	@echo "Total Lines of Code:"
	@find . -name '*.go' -not -path "./vendor/*" | xargs wc -l | tail -1
	@echo "\nNumber of Go files:"
	@find . -name '*.go' -not -path "./vendor/*" | wc -l
	@echo "\nNumber of packages:"
	@find . -type d -not -path "./vendor/*" -not -path "./.git/*" | wc -l

contributors: ## Show contributors
	@echo "$(GREEN)Top Contributors:$(NC)"
	@git log --format='%aN' | sort | uniq -c | sort -rn | head -10

.DEFAULT_GOAL := help