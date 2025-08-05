# Solace REST V2 API Validator MCP Server (Go Implementation)

This is a Go implementation of the Solace REST V2 API Validator MCP server. It provides a more stable and reliable alternative to the Python implementation, with built-in keep-alive mechanisms to prevent timeouts.

## Features

- Validates REST APIs against Solace REST API conventions
- Implements the Model Context Protocol (MCP) for integration with Cline
- Built-in keep-alive mechanism to prevent timeouts
- Graceful shutdown handling
- Comprehensive test suite
- JSON-based validation rules for easy extension
- URL path validation for quick checks without full OpenAPI specs
- Support for all Solace REST API ADRs

## Installation

### Prerequisites

- Go 1.16 or higher
- Make (optional, for using the Makefile)

### Building from Source

1. Clone the repository:

```bash
git clone https://github.com/solace/restv2-api-server-go.git
cd restv2-api-server-go
```

2. Build the server:

```bash
make build
```

This will create a binary in the `build` directory.

### Installing for Cline

To install the server for use with Cline:

```bash
make install-cline-config
```

This will install the Cline MCP settings file in the appropriate location.

## Usage

### Running the Server

To run the server:

```bash
make run
```

By default, the server listens on port 9090. You can specify a different port using the `--port` flag:

```bash
./build/restv2-api-server-go --port 8080
```

To enable the keep-alive mechanism:

```bash
./build/restv2-api-server-go --keep-alive
```

### Environment Variables

- `MCP_PORT`: The port to listen on (default: 9090)
- `MCP_KEEP_ALIVE`: Enable keep-alive mechanism if set to "true"

### Testing

To run the tests:

```bash
make test
```

To test the validator with a sample API:

```bash
make test-validator
```

To test the MCP connection:

```bash
make test-mcp-connection
```

To test the server stability:

```bash
make test-stability
```

To test URL path validation:

```bash
make test-url-path
```

## API

The server implements the following MCP methods:

- `ping`: Tests the server connection
- `getTools`: Returns the available tools
- `getResources`: Returns the available resources
- `validate`: Validates an API specification against a set of rules
- `validateUrlPath`: Validates a URL path against Solace REST API conventions

### Validation Rules

#### Built-in Rules

The server implements the following built-in validation rules:

- `solace_rest_rules`: Validates that the API follows Solace REST API conventions
- `solace_singular_user_resources`: Validates that user-specific resources use `/me/` instead of `/users/{id}`
- `solace_custom_actions`: Validates that custom actions follow Solace conventions

#### JSON-based Rules

The server also supports loading validation rules from JSON files in the `config/rules` directory. These rules implement various Solace REST API ADRs:

- **API Versioning**: Validates API path versioning
- **Resource Naming**: Validates resource naming conventions
- **Collection POST**: Validates collection endpoints have POST methods
- **Audit Fields**: Validates DTOs include standard audit fields
- **Field/Resource Naming**: Validates field and resource naming conventions
- **Payload Structure**: Validates API payload structure
- **Error Responses**: Validates API error responses
- **Pagination**: Validates API pagination
- **Standard Fields**: Validates API standard fields
- **Resource Paths**: Validates API resource paths
- **Delete Behavior**: Validates API DELETE endpoints
- **Sorting**: Validates API sorting
- **Filtering**: Validates API filtering
- **Array Query Parameters**: Validates API array query parameters
- **Long Running Operations**: Validates API long running operations
- **API Deprecation**: Validates API deprecation
- **Time Range Half-Open**: Validates API time range half-open approach
- **Singular User Resources**: Validates singular user resources
- **Enum Naming**: Validates enum values follow UPPER_SNAKE_CASE naming convention

### URL Path Validation

The server now supports validating a single URL path against Solace REST API conventions. This feature allows you to validate a URL path without having to create a complete OpenAPI specification.

When you provide a URL path, the server will:

1. Analyze the URL path structure to extract path parameters
2. Determine if the path ends with a resource or collection
3. Determine appropriate HTTP methods based on the path structure
4. Create a minimal OpenAPI specification for validation
5. Apply the validation rules to the generated specification
6. Return the validation results along with the path analysis

## Extending the Server

### Adding New Rules

To add a new rule:

1. Create a new JSON file in the `config/rules` directory:

```json
{
  "name": "rule_name",
  "description": "Rule description",
  "enabled": true,
  "conditions": [
    {
      "type": "condition_type",
      "pattern": "regex_pattern",
      "message": "Error message"
    }
  ]
}
```

2. The server will automatically load the rule when it starts.

### Adding New Condition Types

To add a new condition type:

1. Modify the `internal/rules/json_rule.go` file to add the new condition type.
2. Implement the condition type's validation logic.

## Troubleshooting

### Server Timeouts

If you experience server timeouts with the Python implementation, try using this Go implementation with the keep-alive mechanism enabled:

```bash
./build/restv2-api-server-go --keep-alive
```

### Cline Integration Issues

If you have issues with Cline integration:

1. Make sure the server is built and installed correctly:

```bash
make build
make install-cline-config
```

2. Restart VS Code to apply the changes.

3. Check the Cline MCP settings file:

```bash
cat ~/Library/Application\ Support/Code/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
