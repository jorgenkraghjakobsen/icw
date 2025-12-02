.PHONY: all build install clean test deps help

# Binary name
BINARY=icw
BUILD_DIR=.
CMD_DIR=cmd/icw

# Version info
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date +%Y-%m-%dT%H:%M:%S%z)
LDFLAGS=-ldflags "-s -w \
	-X github.com/jakobsen/icw/internal/version.Version=$(VERSION) \
	-X github.com/jakobsen/icw/internal/version.Commit=$(COMMIT) \
	-X github.com/jakobsen/icw/internal/version.BuildDate=$(BUILD_DATE)"

# Installation paths
INSTALL_PATH=$(HOME)/bin
COMPLETION_PATH=/usr/local/share/bash-completion/completions

all: build

help:
	@echo "ICW Go Build System"
	@echo "==================="
	@echo "make build         - Build the icw binary"
	@echo "make install       - Install to ~/bin and bash completion"
	@echo "make clean         - Remove built binary"
	@echo "make test          - Run tests"
	@echo "make deps          - Install Go dependencies"

deps:
	go mod download
	go mod tidy

build:
	@echo "Building $(BINARY) $(VERSION) ($(COMMIT))..."
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./$(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY)"

install: build
	@echo "Installing $(BINARY) to $(INSTALL_PATH)..."
	mkdir -p $(INSTALL_PATH)
	cp $(BUILD_DIR)/$(BINARY) $(INSTALL_PATH)/
	@if [ -f completions/icw_bashcompletion.sh ]; then \
		echo "Installing bash completion..."; \
		sudo mkdir -p $(COMPLETION_PATH); \
		sudo cp completions/icw_bashcompletion.sh $(COMPLETION_PATH)/$(BINARY); \
	fi
	@echo "Installation complete!"
	@echo "Run '$(BINARY) --version' to verify"

clean:
	@echo "Cleaning..."
	rm -f $(BUILD_DIR)/$(BINARY)

test:
	@echo "Running tests..."
	go test -v ./...
