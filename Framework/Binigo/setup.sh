#!/bin/bash

# Binigo Framework Repository Setup Script
# This script sets up a complete Binigo repository ready for publishing

set -e

REPO_NAME="binigo"
GITHUB_USER=""
PROJECT_DESC="A Laravel-inspired web framework for Go built on FastHTTP"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
cat << "EOF"
  ____  _       _             
 |  _ \(_)_ __ (_) __ _  ___  
 | |_) | | '_ \| |/ _` |/ _ \ 
 |  _ <| | | | | | (_| | (_) |
 |_| \_\_|_| |_|_|\__, |\___/ 
                  |___/        
 Framework Setup
EOF
echo -e "${NC}"

# Check if GitHub CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${YELLOW}GitHub CLI not found. Install from: https://cli.github.com${NC}"
    echo "Continue without creating GitHub repo? (y/n)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        exit 1
    fi
    SKIP_GITHUB=true
fi

# Get GitHub username
if [ -z "$SKIP_GITHUB" ]; then
    echo -e "${GREEN}Enter your GitHub username:${NC}"
    read -r GITHUB_USER
    
    if [ -z "$GITHUB_USER" ]; then
        echo -e "${RED}GitHub username is required${NC}"
        exit 1
    fi
fi

# Create project directory
echo -e "${GREEN}Creating project structure...${NC}"
mkdir -p $REPO_NAME
cd $REPO_NAME

# Create directory structure
mkdir -p cmd/binigo
mkdir -p pkg/{router,context,middleware,database,validation,container}
mkdir -p internal/utils
mkdir -p examples/{blog-api,todo-app,ecommerce}
mkdir -p docs
mkdir -p tests/integration
mkdir -p scripts
mkdir -p .github/workflows
mkdir -p .github/ISSUE_TEMPLATE

# Create go.mod
echo -e "${GREEN}Initializing Go module...${NC}"
cat > go.mod << EOF
module github.com/$GITHUB_USER/binigo

go 1.21

require (
	github.com/valyala/fasthttp v1.51.0
	github.com/lib/pq v1.10.9
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
)
EOF

# Create .gitignore
echo -e "${GREEN}Creating .gitignore...${NC}"
cat > .gitignore << 'EOF'
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test binary
*.test
*.out

# Dependencies
vendor/

# Go workspace
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.*
!.env.example

# Logs
*.log
storage/logs/*
!storage/logs/.gitkeep

# Uploads
storage/uploads/*
!storage/uploads/.gitkeep

# Coverage
coverage.out
coverage.html
EOF

# Create .env.example
cat > .env.example << 'EOF'
APP_NAME=Binigo
APP_ENV=development
APP_DEBUG=true
APP_PORT=8080

DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=binigo
DB_USERNAME=postgres
DB_PASSWORD=
EOF

# Create basic package files
echo -e "${GREEN}Creating package structure...${NC}"

# Create binigo.go (main export file)
cat > binigo.go << 'EOF'
package binigo

// Re-export main types for convenience
type (
	Application  = application
	Router       = router
	Context      = context
	HandlerFunc  = handlerFunc
	MiddlewareFunc = middlewareFunc
	Container    = container
	Config       = config
	DB           = database
	Map          = map[string]interface{}
)

// New creates a new Binigo application
func New(cfg *Config) *Application {
	return newApplication(cfg)
}

// Bootstrap creates an application with default configuration
func Bootstrap() *Application {
	return New(LoadConfig())
}

// Version returns the framework version
func Version() string {
	return "1.0.0"
}
EOF

# Create GitHub Actions workflow
echo -e "${GREEN}Creating CI/CD workflows...${NC}"
cat > .github/workflows/ci.yml << 'EOF'
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test -v -race -coverprofile=coverage.out ./...
      - uses: codecov/codecov-action@v3

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - uses: golangci/golangci-lint-action@v3
EOF

# Create SECURITY.md
cat > SECURITY.md << 'EOF'
# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability, please email security@binigo.dev

Please do not report security vulnerabilities through public GitHub issues.
EOF

# Create CHANGELOG.md
cat > CHANGELOG.md << 'EOF'
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2024-01-01

### Added
- Initial release
- Core framework functionality
- Router with middleware support
- Database query builder
- Request validation
- Service container
- CLI tool
EOF

# Initialize git
echo -e "${GREEN}Initializing git repository...${NC}"
git init
git add .
git commit -m "🎉 Initial commit: Binigo Framework"

# Create GitHub repository if CLI is available
if [ -z "$SKIP_GITHUB" ]; then
    echo -e "${GREEN}Creating GitHub repository...${NC}"
    gh repo create $REPO_NAME --public --description "$PROJECT_DESC" --source=. --remote=origin --push
    
    echo -e "${GREEN}Setting up repository settings...${NC}"
    gh repo edit --enable-issues --enable-discussions --enable-wiki
    
    echo -e "${GREEN}Creating first release...${NC}"
    git tag v1.0.0
    git push origin v1.0.0
    
    echo -e "${GREEN}Repository created: https://github.com/$GITHUB_USER/$REPO_NAME${NC}"
fi

# Success message
echo ""
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  ✓ Binigo Framework Setup Complete!   ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo -e "  1. Review and update README.md"
echo -e "  2. Add your framework code to pkg/"
echo -e "  3. Implement cmd/binigo/main.go"
echo -e "  4. Write tests"
echo -e "  5. Update documentation"
echo ""
echo -e "${BLUE}Useful commands:${NC}"
echo -e "  make build          - Build the CLI"
echo -e "  make test           - Run tests"
echo -e "  make lint           - Run linters"
echo -e "  make install        - Install CLI locally"
echo ""
echo -e "${YELLOW}Don't forget to:${NC}"
echo -e "  - Update GitHub username in go.mod"
echo -e "  - Add topics to GitHub repository"
echo -e "  - Enable GitHub Pages for documentation"
echo -e "  - Set up branch protection rules"
echo ""
echo -e "${GREEN}Happy coding! 🚀${NC}"