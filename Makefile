# Makefile for Ambros - The Command Butler

# Variables
BINARY_NAME=ambros
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
DIST_DIR=dist

# Linker flags for versioning
LDFLAGS=-ldflags "-X 'github.com/gi4nks/ambros/v3/cmd/commands.Version=${VERSION}' \
		 -X 'github.com/gi4nks/ambros/v3/cmd/commands.GitCommit=${COMMIT}' \
		 -X 'github.com/gi4nks/ambros/v3/cmd/commands.BuildDate=${DATE}' \
		 -s -w"

# Build Targets
.PHONY: all build run clean tidy test install dist docker

all: build

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) main.go

run: build
	./$(BINARY_NAME)

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -rf $(DIST_DIR) bin coverage

tidy:
	go mod tidy

test:
	go test -v ./...

test-integration:
	go test -v ./tests/integration/...

test-all: test test-integration

# Cross-compilation targets for distribution
dist: clean
	mkdir -p $(DIST_DIR)
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	# Create checksums for each binary
	cd $(DIST_DIR) && sha256sum * > checksums.txt

# Docker image build
docker:
	docker build -t ambros:$(VERSION) .
	docker tag ambros:$(VERSION) ambros:latest

# Local installation
install: build
	go install .
