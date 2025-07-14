.PHONY: build test lint clean install run help

# Variables
BINARY_NAME=mde
GO=go
GOLINT=golangci-lint
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOGET=$(GO) get
GOMOD=$(GO) mod
INSTALL_PATH=/usr/local/bin

# Default target
help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  test      - Run tests"
	@echo "  lint      - Run linters"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Install binary to $(INSTALL_PATH)"
	@echo "  run       - Run the application"

build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/mde

test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

lint:
	$(GOLINT) run ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

install: build
	cp $(BINARY_NAME) $(INSTALL_PATH)

run: build
	./$(BINARY_NAME)

# Development helpers
.PHONY: deps fmt vet

deps:
	$(GOMOD) download
	$(GOMOD) tidy

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...