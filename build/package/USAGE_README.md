# Solace REST V2 MCP Go Server Usage Guide

This guide explains how to use the Solace REST V2 MCP Go server with Cline to validate REST APIs against Solace conventions.

## Installation

### Prerequisites

- Go 1.16 or higher
- Make (optional, for using the Makefile)
- VS Code with Cline extension installed

### Building and Installing

1. Clone the repository:

```bash
git clone https://github.com/solacedev/restv2-api-server-go.git
cd restv2-api-server-go
```

2. Build the server:

```bash
make build
```

3. Install the Cline MCP settings:

```bash
make install-cline-config
```

4. Restart VS Code for the changes to take effect.

## Using the MCP Server with Cline

### Cline Integration

For Cline to use the Solace REST V2 MCP Go server, it needs to be configured in the Cline MCP settings file. The `make install-cline-config` command installs the necessary configuration:

```json
{
  "mcpServers": {
    "Solace REST V2 MCP Go": {
      "autoApprove": [],
      "disabled": false,
      "timeout": 300,
      "type": "stdio",
      "command": "/path/to/restv2-api-server-go",
      "cwd": "/path/to/project",
      "env": {
        "MCP_KEEP_ALIVE": "true",
        "MCP_PORT": "9090"
      }
    }
  }
}
```

This configuration tells Cline:
1. The name of the MCP server ("Solace REST V2 MCP Go")
2. The command to run the server
3. The working directory
4. Environment variables to set

When you ask Cline to validate a REST API or URL path, it will:
1. Start the MCP server if it's not already running
2. Send the validation request to the server
3. Return the results to you

### Starting a Conversation

1. Open VS Code with the Cline extension.
2. Start a new conversation with Cline.
3. Ask Cline to validate your REST API against Solace conventions.

### Sample Prompts

#### Basic Validation

```
Can you validate this REST API against Solace conventions?

openapi: 3.0.0
info:
  title: My API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get all users
      responses:
        '200':
          description: OK
  /users/{id}:
    get:
      summary: Get user by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
```

#### Validation of a Single URL Path

To trigger the URL path validation feature in Cline, use one of these specific prompt formats:

```
Can you validate this REST API endpoint against Solace conventions?
/api/v0/admin/cloudAgents/{datacenterId}/upgrades
```

Or:

```
Please validate this URL path against Solace REST conventions:
/api/v0/admin/cloudAgents/{datacenterId}/upgrades
```

The key phrases that Cline recognizes are:
- "validate this REST API endpoint"
- "validate this URL path"

When Cline recognizes these phrases followed by a URL path, it will:
1. Analyze the URL path structure to extract all necessary information
2. Automatically determine appropriate HTTP methods based on REST conventions
   - GET for both collection and resource paths
   - POST for collection paths (paths not ending with an ID parameter)
   - PUT, PATCH, DELETE for resource paths (paths ending with an ID parameter)
3. Create a minimal OpenAPI specification for validation
4. Send it to the MCP server for validation
5. Return the validation results with recommendations

The MCP server will automatically detect that:
- The path `/api/v0/admin/cloudAgents/{datacenterId}/upgrades` contains a path parameter `datacenterId`
- The path ends with a collection resource `upgrades`
- Based on REST conventions, GET and POST would be appropriate for this path

You only need to provide the URL path - the MCP server handles the analysis and validation automatically.

#### Validation with Specific Rules

```
Can you validate this API against the solace_singular_user_resources rule?

openapi: 3.0.0
info:
  title: User API
  version: 1.0.0
paths:
  /users/{id}/profile:
    get:
      summary: Get user profile
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
```

#### Fixing Validation Issues

```
I received these validation issues for my API. Can you help me fix them?

Issues:
- POST should be used for collection paths, not for specific resources
- User-specific resources should use /me/ instead of /users/{id}

Here's my API:
openapi: 3.0.0
info:
  title: Problem API
  version: 1.0.0
paths:
  /users/{id}:
    post:
      summary: Update user
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
```

#### Analyzing an Existing API

```
Can you analyze this API and tell me if it follows Solace REST conventions?

openapi: 3.0.0
info:
  title: E-commerce API
  version: 1.0.0
paths:
  /products:
    get:
      summary: List products
      responses:
        '200':
          description: OK
    post:
      summary: Create product
      responses:
        '201':
          description: Created
  /products/{id}:
    put:
      summary: Update product
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
```

### How Cline Uses the MCP Server

When you provide an API specification to Cline, it:

1. Extracts the OpenAPI specification from your message
2. Sends it to the Solace REST V2 MCP Go server for validation
3. Applies the validation rules (either all rules or specific ones you requested)
4. Returns the validation results with detailed explanations
5. Can provide suggestions for fixing any issues found

The MCP server handles the technical validation while Cline provides a user-friendly interface and explanations.

### Available Validation Rules

#### Built-in Rules

The server implements the following built-in validation rules:

1. **solace_rest_rules**: Validates that the API follows Solace REST API conventions
   - Checks if HTTP methods are appropriate for the path (e.g., POST for collections, PUT/PATCH/DELETE for resources)

2. **solace_singular_user_resources**: Validates that user-specific resources use `/me/` instead of `/users/{id}`
   - Ensures that user-specific resources follow the convention of using `/me/` for the current user

3. **solace_custom_actions**: Validates that custom actions follow Solace conventions
   - Verifies that custom actions use the POST method and follow the `/actions/` pattern

#### JSON-based Rules

In addition to the built-in rules, the server supports loading validation rules from JSON files. These files are stored in the `config/rules` directory and are loaded automatically when the server starts.

Each JSON rule file should have the following structure:

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

The following condition types are supported:

1. **path_pattern**: Validates that paths match a regex pattern
   ```json
   {
     "type": "path_pattern",
     "pattern": "^/api/v[0-9]+/.*$",
     "message": "API paths must start with /api/v{number}/"
   }
   ```

2. **method_check**: Validates that a specific path has a specific method
   ```json
   {
     "type": "method_check",
     "path": "/api/v1/users",
     "method": "post",
     "message": "Collection endpoint should support POST method"
   }
   ```

3. **parameter_check**: Validates that a specific path has parameters
   ```json
   {
     "type": "parameter_check",
     "path": "/api/v1/users/{id}",
     "message": "Path should have parameters"
   }
   ```

4. **resource_naming**: Validates that resource names match a regex pattern
   ```json
   {
     "type": "resource_naming",
     "pattern": "^[a-z][a-zA-Z0-9]*$",
     "message": "Resource names must start with a lowercase letter"
   }
   ```

5. **schema_field**: Validates that schemas include specific fields
   ```json
   {
     "type": "schema_field",
     "field": "createdTime",
     "format": "date-time",
     "message": "DTOs must include 'createdTime' field in ISO 8601 format"
   }
   ```

Example JSON rule files are provided in the `config/rules` directory:

- **api_versioning.json**: Validates that API paths include proper versioning as per ADR:
  - API paths must start with /api/v{number}/ (e.g., /api/v2/)
  - A new API version is a major change that implies fundamental changes to the API
  - Only backwards compatible and emergency changes will be made to stable APIs within a version
  - An emergency change is a change made for security, regulatory, or specification violation reasons
  - A breaking change to a resource would require a whole new resource
  - A major change to an entire product area would require a new product area
- **resource_naming.json**: Validates that resource names follow proper naming conventions
- **collection_post.json**: Validates that collection endpoints have POST methods
- **audit_fields.json**: Validates that DTOs include standard audit fields as per ADR
- **field_resource_naming.json**: Validates that field and resource naming follows Solace REST API conventions as per ADR:
  - URL structure must start with `/api/v2/<product area>`
  - Collections must be represented by plural nouns
  - IDs that refer to other resources are prefixed with the type of object
  - The resource being accessed should only use `id` as the field
  - The path must not include an organization ID variable
  - Camel case for resources and fields
  - No shortened words (with some exceptions)
  - Common abbreviations are allowed and should use camel case
- **payload_structure.json**: Validates that API payload structure follows Solace REST API conventions as per ADR:
  - All responses must include an envelope containing a data field for the resulting objects
  - Requests do not use a data envelope
  - Metadata about the response (ex. pagination) will be in a meta JSON object
- **error_responses.json**: Validates that API error responses follow Solace REST API conventions as per ADR:
  - Errors must include a message, error ID, meta, and validation details
  - The message is a user-friendly message detailing what went wrong
  - The errorId is a UUID that was also logged with an appropriate stack trace
  - The validationDetails describe what the issue was with the fields
  - The field name must match the field name being set in the request payload
  - If the field is an object, the validationDetails can include nested fields
  - If the field is an array, the array index must be indicated with square brackets
  - All validation errors should be returned in a single response
- **pagination.json**: Validates that API pagination follows Solace REST API conventions as per ADR:
  - REST APIs should accept pageSize and pageNumber request parameters
  - Responses should include pagination information in the meta.pagination object with:
    - count: the total number of elements available across all pages
    - pageNumber: the current page number being returned (starts at 1, not 0)
    - pageSize: the page size requested by the user or the default page size
    - nextPage: the next page available with results (null if there are no other pages)
    - totalPages: the total number of pages
- **standard_fields.json**: Validates that API resources include standard fields as per ADR:
  - Every REST resource must have an `id` field that represents the opaque ID of the object
  - Every REST resource must have a `type` field that uniquely identifies the type of object being returned
- **resource_paths.json**: Validates that API resource paths follow Solace REST API conventions as per ADR:
  - REST APIs should avoid forcing the user to provide non-identifying attributes in the path
  - Resource paths should follow the pattern /api/v{number}/{product area}/{resource type}/{id}
  - Non-identifying attributes should be provided in the request body or as query parameters
- **delete_behavior.json**: Validates that API DELETE endpoints follow Solace REST API conventions as per ADR:
  - DELETE on a non-existent resource should return a 404 Not Found
  - Successful DELETE with an entity describing the status should return a 200 OK
  - Successful DELETE for an action that has been queued should return a 202 Accepted
  - Successful DELETE without an entity in the response should return a 204 No Content
- **sorting.json**: Validates that API sorting follows Solace REST API conventions as per ADR:
  - REST APIs should accept the 'sort' request parameter for sorting
  - Sort parameter value should be either a field name or a field name and direction delimited by a colon
  - Sort direction should be 'asc' (default) or 'desc'
  - Examples: '?sort=name', '?sort=name:asc', '?sort=name:desc'
- **filtering.json**: Validates that API filtering follows Solace REST API conventions as per ADR:
  - Filter that applies to a property of the object MUST match the name of the object property
  - Filter query parameters MUST follow SEMPv2's syntax if operators are to be introduced (e.g. '==' | '!=' | '<' | '>' | '<=' | '>=')
  - When filtering by multiple key/value pairs, each filter is separated by a semicolon (';') for AND semantic and comma (',') for OR semantic
  - Special characters (e.g. ',', ';') need to be escaped in the query
  - Examples: '?colour==red', '?colour==red;security==high', '?colour==red,green,blue;security==high'
- **array_query_parameters.json**: Validates that API array query parameters follow Solace REST API conventions as per ADR:
  - Array query parameters should be passed as comma-delimited strings
  - Document that the parameter holds a comma-delimited string using the Swagger description annotation
  - Document minimum and maximum array size where appropriate
  - Examples: '?ids=string1,string2,string3', '?tags=tag1,tag2,tag3', '?environmentIds=env-123,env-456,env-789'
- **long_running_operations.json**: Validates that API long running operations follow Solace REST API conventions as per ADR:
  - HTTP status 202 MUST be returned for long running operations
  - The response SHOULD contain a location header with the Operation resource URI
  - Operations must be available as a sub-resource of the affected resource
  - Operation resource must include id, operationType, createdBy, createdTime, and status fields
  - The minimum states that MUST be supported are: pending, inProgress, succeeded, failed
  - Error object MUST exist when status is 'failed' and include message and errorId fields
  - If parallel or queuing operations are not supported, HTTP status 409 (Conflict) MUST be returned
- **api_deprecation.json**: Validates that API deprecation follows Solace REST API conventions as per ADR:
  - The part of the API being deprecated MUST be indicated in the documentation
  - Deprecation description MUST be indicated in the documentation
  - Deprecation description MUST include reason for deprecation, replacement if any, and proposed date of removal
  - The response of the deprecated API MUST include a X-Solace-API-Deprecated header with the value being a link to the documentation
  - If generally available API/functionality is being replaced, the replacement MUST also be GA before deprecation
  - Date of removal MUST be stated in ISO8601 format
  - For deprecated operations/endpoints, deprecation description must be in the summary
  - For deprecated parameters, deprecation description must be in the description
  - For deprecated request body properties, deprecation description must be in the description
- **time_range_half_open.json**: Validates that API time ranges follow the Half-Open approach as per ADR:
  - An API with a time range MUST use the Half-Open approach
  - The start time is always inclusive while the end time is always exclusive
  - The API documentation must clearly state that the start time is inclusive and the end time is exclusive
  - Time range parameters should follow naming conventions (e.g., startTime/endTime, fromDate/toDate)
  - Time range parameters should use ISO8601 format for consistency
- **singular_user_resources.json**: Validates that single, unique resources owned by the currently logged in user use singular nouns as per ADR:
  - Use singular nouns for retrieving a resource associated with the current security context
  - This applies to resources that have a 1-1 relationship with an attribute that is retrievable from the security context
  - Example: use `user` instead of `users/{id}` for the current user
  - API documentation should clearly state that 'user' refers to the currently logged in user
- **enum_naming.json**: Validates that enum values follow UPPER_SNAKE_CASE naming convention as per ADR:
  - Enum values MUST be UPPER_SNAKE_CASE
  - Exception: Enum values inherited from other APIs SHOULD match their style for consistency (ex. SEMPv2 enums like non-exclusive being kebab-case)
  - If enum values are inherited from other APIs, they should be documented as such

### URL Path Validation

The server now supports validating a single URL path against Solace REST API conventions. This feature allows you to validate a URL path without having to create a complete OpenAPI specification.

When you provide a URL path, the server will:

1. Analyze the URL path structure to extract path parameters
2. Determine if the path ends with a resource or collection
3. Determine appropriate HTTP methods based on the path structure
4. Create a minimal OpenAPI specification for validation
5. Apply the validation rules to the generated specification
6. Return the validation results along with the path analysis

#### Implementation Details

The URL path validation feature is implemented in the following files:

- `internal/validator/url_path_validator.go`: Contains the URL path validator implementation
- `internal/validator/validator.go`: Adds a method for validating URL paths
- `internal/server/server.go`: Adds support for the new method and tool

The implementation follows these steps:

1. Extract path parameters from the URL path (e.g., `{datacenterId}` from `/api/v0/admin/cloudAgents/{datacenterId}/upgrades`)
2. Determine if the path ends with a resource (ID parameter) or collection
3. Determine appropriate HTTP methods based on the path structure:
   - GET is valid for both collection and resource paths
   - POST is valid for collection paths (paths not ending with an ID parameter)
   - PUT, PATCH, DELETE are valid for resource paths (paths ending with an ID parameter)
4. Create a minimal OpenAPI specification for validation
5. Apply the validation rules to the generated specification
6. Return the validation results along with the path analysis

### Example Validation Results

When you submit an API for validation, Cline will return results like this:

```json
{
  "results": {
    "solace_rest_rules": {
      "status": "passed"
    },
    "solace_singular_user_resources": {
      "status": "failed",
      "issues": [
        {
          "path": "/users/{id}/profile",
          "message": "User-specific resources should use /me/ instead of /users/{id}"
        }
      ]
    },
    "solace_custom_actions": {
      "status": "passed"
    }
  }
}
```

### Fixing Validation Issues

Based on the validation results, you can make changes to your API specification to address any issues:

1. For `solace_rest_rules` issues:
   - Use GET for both collection and resource paths
   - Use POST for collection paths (e.g., `/users`)
   - Use PUT, PATCH, DELETE for resource paths (e.g., `/users/{id}`)

2. For `solace_singular_user_resources` issues:
   - Replace `/users/{id}` with `/me` for endpoints that refer to the current user

3. For `solace_custom_actions` issues:
   - Ensure custom actions use the POST method
   - Follow the pattern `/resources/{id}/actions/action-name`

## Advanced Usage

### Running the Server Manually

If you need to run the server manually:

```bash
./build/restv2-api-server-go --keep-alive
```

By default, the server listens on port 9090. You can specify a different port:

```bash
./build/restv2-api-server-go --port 8080
```

### Environment Variables

- `MCP_PORT`: The port to listen on (default: 9090)
- `MCP_KEEP_ALIVE`: Enable keep-alive mechanism if set to "true"

### Testing the Server

You can test the server using the provided example scripts:

1. Test MCP Connection:
   ```bash
   make test-mcp-connection
   ```

2. Test Validator with Sample API:
   ```bash
   make test-validator
   ```

3. Test URL Path Validation:
   ```bash
   make test-url-path
   ```

4. Test Server Stability:
   ```bash
   make test-stability
   ```

## Troubleshooting

### Server Timeouts

If you experience server timeouts:

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

4. Ensure the paths in the settings file are correct:
   - `command` should point to the built binary
   - `cwd` should point to the project directory

## Direct API Validation (Without Cline)

You can also validate APIs directly using the validator:

```bash
go run examples/test_validator.go examples/sample-api.yaml
```

This will output the validation results for the sample API.
