BINARY_NAME := config-analyzer
BUILD_DIR   := ./bin

.PHONY: all build clean test run vet fmt proto tidy

## Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

## Run tests with race detection and coverage
test:
	go test -v -race -cover ./...

## Run go vet on all packages
vet:
	go vet ./...

## Run gofmt
fmt:
	gofmt -s -w ./

## Tidy dependencies
tidy:
	go mod tidy

## Generate protobuf code (requires protoc + protoc-gen-go + protoc-gen-go-grpc)
proto:
	protoc --go_out=. --go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		internal/grpc/proto/service.proto

## Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

## Rebuild from scratch
all: clean build test

# ── Convenience targets ─────────────────────────────────────────────

# Run CLI:  make cli ARGS="analyze configs/config.json"
cli: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Run REST server
rest: build
	@$(BUILD_DIR)/$(BINARY_NAME) serve-http

# Run gRPC server
grpc: build
	@$(BUILD_DIR)/$(BINARY_NAME) serve-grpc
