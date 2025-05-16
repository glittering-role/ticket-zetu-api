# The Go binary name
BINARY_NAME=h-api-go

# Directories
SRC_DIR=./cmd

# Default port and configuration
PORT=8080
CONFIG_PATH=.env

# Go commands
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_RUN=$(GO_CMD) run
GO_FMT=$(GO_CMD) fmt
GO_TEST=$(GO_CMD) test
GO_GET=$(GO_CMD) get

# Set default target
.PHONY: all
all: run

# Run the application
.PHONY: run
run:
	$(GO_RUN) $(SRC_DIR)/main.go

# Build the application
.PHONY: build
build:
	$(GO_BUILD) -o $(BINARY_NAME) $(SRC_DIR)/main.go

# Format Go code
.PHONY: fmt
fmt:
	$(GO_FMT) ./...

# Test the application
.PHONY: test
test:
	$(GO_TEST) -v ./...

# Clean generated files
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
