# Variables
BINARY_NAME=scanexpress
INSTALL_PATH?=$(HOME)/.local/bin
DIST_DIR=dist

.PHONY: help run build clean install uninstall all

# Default goal
.DEFAULT_GOAL := help

help:
	@echo "Available commands:"
	@echo "  make help        Show this help message"
	@echo "  make run         Run the application"
	@echo "  make build       Build the application binary"
	@echo "  make clean       Clean the build directory"
	@echo "  make install     Install the application"
	@echo "                   Override install path: make install INSTALL_PATH=/path/to/install"
	@echo "  make uninstall   Uninstall the application"
	@echo "                   Specify install path if changed: make uninstall INSTALL_PATH=/path/to/install"
	@echo "  make all         Build the application (same as make build)"


all: build

# Run the application
run:
	@echo "Running application..."
	@go run main.go

# Build the application
build:
	@echo "Building application..."
	@mkdir -p $(DIST_DIR) # Ensure dist directory exists
	@CGO_ENABLED=0 go build -ldflags='-w -s' -o $(DIST_DIR)/$(BINARY_NAME) main.go
	@echo "Binary built: $(DIST_DIR)/$(BINARY_NAME)"

# Clean the build directory
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(DIST_DIR)
	@echo "Build directory cleaned."

# Install the application
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@install -d $(INSTALL_PATH) # Ensure install directory exists
	@install -m 755 $(DIST_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed successfully to $(INSTALL_PATH)/$(BINARY_NAME)."

# Uninstall the application
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_PATH)..."
	@rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "$(BINARY_NAME) uninstalled successfully from $(INSTALL_PATH)."
