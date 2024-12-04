# Build configuration
ROCM_PATH ?= /opt/rocm
GO ?= go
GOFLAGS ?=
LDFLAGS ?=

# Export CGO settings
export CGO_ENABLED=1
export CGO_CFLAGS=-I${ROCM_PATH}/include
export CGO_LDFLAGS=-L${ROCM_PATH}/lib -lrocm_smi64

# Binary names
EXAMPLE_BIN=device_info

# Directories
PKG_DIR=pkg
EXAMPLES_DIR=examples
BUILD_DIR=build

.PHONY: all
all: build example

.PHONY: build
build:
	$(GO) build $(GOFLAGS) ./...

.PHONY: example
example:
	mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(EXAMPLE_BIN) $(EXAMPLES_DIR)/main.go

.PHONY: test
test:
	$(GO) test $(GOFLAGS) ./...

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	$(GO) clean

.PHONY: vendor
vendor:
	$(GO) mod vendor
	$(GO) mod tidy

.PHONY: check
check:
	$(GO) vet ./...
	$(GO) fmt ./...

.PHONY: install
install:
	$(GO) install $(GOFLAGS) ./...

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all      - Build everything (default)"
	@echo "  build    - Build the library"
	@echo "  example  - Build the example program"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  vendor   - Update vendor directory"
	@echo "  check    - Run code checks"
	@echo "  install  - Install the library"
	@echo "  help     - Show this help"