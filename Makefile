APP_NAME := editpilot
GO ?= $(HOME)/.go/bin/go
BINARY_DIR := bin
BINARY_PATH := $(BINARY_DIR)/$(APP_NAME)
INSTALL_DIR ?= $(HOME)/.local/bin

.PHONY: help build install uninstall test run clean fmt tidy

help:
	@echo "Available targets:"
	@echo "  make build      - Build the editpilot binary into ./bin"
	@echo "  make install    - Install the binary into $(INSTALL_DIR)"
	@echo "  make uninstall  - Remove installed binary from $(INSTALL_DIR)"
	@echo "  make test       - Run all tests"
	@echo "  make run        - Run the CLI via go run"
	@echo "  make fmt        - Format Go code"
	@echo "  make tidy       - Run go mod tidy"
	@echo "  make clean      - Remove build artifacts"

build:
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(BINARY_PATH) ./cmd/editpilot

install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(BINARY_PATH) $(INSTALL_DIR)/$(APP_NAME)
	@echo "Installed $(APP_NAME) to $(INSTALL_DIR)/$(APP_NAME)"

test:
	$(GO) test ./...

run:
	$(GO) run ./cmd/editpilot

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BINARY_DIR)

uninstall:
	rm -f $(INSTALL_DIR)/$(APP_NAME)
	@echo "Removed $(INSTALL_DIR)/$(APP_NAME)"
