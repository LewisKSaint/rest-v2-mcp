GO=go
BINARY_NAME=restv2-api-server-go
BUILD_DIR=build

.PHONY: all build clean run test install test-validator test-mcp-connection test-stability test-url-path test-audit-fields test-enum-naming test-singular-user-resources package install-cline-config

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean

run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

test:
	@echo "Running tests..."
	$(GO) test -v ./...

install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

test-validator: build
	@echo "Testing validator with sample API..."
	$(GO) run ./examples/test_validator.go ./examples/sample-api.yaml

test-mcp-connection: build
	@echo "Testing MCP connection..."
	@$(BUILD_DIR)/$(BINARY_NAME) & \
	PID=$$!; \
	sleep 2; \
	$(GO) run ./examples/test_mcp_connection.go; \
	kill $$PID

test-stability: build
	@echo "Testing server stability..."
	@$(BUILD_DIR)/$(BINARY_NAME) & \
	PID=$$!; \
	sleep 2; \
	$(GO) run ./examples/test_server_stability.go --duration 60 --interval 5; \
	kill $$PID

test-url-path: build
	@echo "Testing URL path validation..."
	@$(BUILD_DIR)/$(BINARY_NAME) & \
	PID=$$!; \
	sleep 2; \
	$(GO) run ./examples/test_url_path_validator.go; \
	kill $$PID

test-audit-fields: build
	@echo "Testing audit fields validation..."
	@$(BUILD_DIR)/$(BINARY_NAME) & \
	PID=$$!; \
	sleep 2; \
	$(GO) run ./examples/test_audit_fields.go ./examples/sample-api-with-schema.yaml; \
	echo "Testing with missing audit fields (should fail)..."; \
	$(GO) run ./examples/test_audit_fields.go ./examples/sample-api-missing-audit-fields.yaml || echo "Failed as expected"; \
	kill $$PID

test-enum-naming: build
	@echo "Testing enum naming validation..."
	@$(BUILD_DIR)/$(BINARY_NAME) & \
	PID=$$!; \
	sleep 2; \
	$(GO) run ./examples/test_validator.go ./examples/sample-api-enum-naming.yaml; \
	kill $$PID

test-singular-user-resources: build
	@echo "Testing singular user resources validation..."
	@$(BUILD_DIR)/$(BINARY_NAME) & \
	PID=$$!; \
	sleep 2; \
	$(GO) run ./examples/test_validator.go ./examples/sample-api-singular-user-resources.yaml; \
	kill $$PID

package: build
	@echo "Packaging $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)/package
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/package/
	@cp -r config $(BUILD_DIR)/package/
	@cp README.md USAGE_README.md $(BUILD_DIR)/package/
	@cp install_cline_config.sh $(BUILD_DIR)/package/
	@cp cline_mcp_settings.json $(BUILD_DIR)/package/
	@cd $(BUILD_DIR) && tar -czf $(BINARY_NAME).tar.gz package
	@echo "Package created at $(BUILD_DIR)/$(BINARY_NAME).tar.gz"

install-cline-config:
	@echo "Installing Cline MCP settings..."
	@mkdir -p ~/Library/Application\ Support/Code/User/globalStorage/saoudrizwan.claude-dev/settings/
	@cp cline_mcp_settings.json ~/Library/Application\ Support/Code/User/globalStorage/saoudrizwan.claude-dev/settings/
