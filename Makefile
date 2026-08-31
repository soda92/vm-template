BINARY_NAME=vm-template
BIN_DIR=bin
GOOS?=linux
GOARCH?=amd64

.PHONY: all build build-static clean test deploy

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/vm-template

build-static:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -ldflags="-s -w -extldflags '-static'" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/vm-template

clean:
	rm -rf $(BIN_DIR)

test:
	go test -v ./...

# Example: make deploy VM_HOST=192.168.1.89 VM_USER=vmware
deploy: build-static
	@if [ -z "$(VM_HOST)" ]; then \
		echo "Usage: make deploy VM_HOST=<ip> [VM_USER=<user>]"; \
		exit 1; \
	fi
	scp $(BIN_DIR)/$(BINARY_NAME) $(or $(VM_USER),vmware)@$(VM_HOST):/tmp/$(BINARY_NAME)
	@echo "Binary copied to /tmp/$(BINARY_NAME) on $(VM_HOST)"
	@echo "Run on target: sudo /tmp/$(BINARY_NAME) inspect"
	@echo "Execute sanitization: sudo /tmp/$(BINARY_NAME) prepare --poweroff"
