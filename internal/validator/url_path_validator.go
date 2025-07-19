package validator

import (
	"fmt"
	"strings"
)

// URLPathValidator validates a single URL path against Solace REST API conventions
type URLPathValidator struct {
	validator *Validator
}

// NewURLPathValidator creates a new URLPathValidator instance
func NewURLPathValidator(validator *Validator) *URLPathValidator {
	return &URLPathValidator{
		validator: validator,
	}
}

// ValidateURLPath validates a single URL path against Solace REST API conventions
func (v *URLPathValidator) ValidateURLPath(urlPath string) (map[string]interface{}, error) {
	// Extract path parameters
	pathParams := v.extractPathParams(urlPath)

	// Determine if the path ends with a resource or collection
	endsWithResource := v.endsWithResource(urlPath)

	// Determine appropriate HTTP methods
	methods := v.determineHTTPMethods(endsWithResource)

	// Create a minimal OpenAPI specification
	spec := v.createOpenAPISpec(urlPath, pathParams, methods)

	// Validate the OpenAPI specification
	params := map[string]interface{}{
		"api_spec": spec,
	}

	results, err := v.validator.Validate(params)
	if err != nil {
		return nil, fmt.Errorf("error validating URL path: %v", err)
	}

	// Add URL path analysis to the results
	analysisResults := map[string]interface{}{
		"path_analysis": map[string]interface{}{
			"path":                urlPath,
			"path_parameters":     pathParams,
			"ends_with_resource":  endsWithResource,
			"appropriate_methods": methods,
		},
	}

	// Merge the validation results with the analysis results
	for k, v := range results {
		analysisResults[k] = v
	}

	return analysisResults, nil
}

// extractPathParams extracts path parameters from the URL path
func (v *URLPathValidator) extractPathParams(urlPath string) []string {
	var params []string
	segments := strings.Split(urlPath, "/")

	for _, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			// Extract parameter name without braces
			paramName := segment[1 : len(segment)-1]
			params = append(params, paramName)
		}
	}

	return params
}

// endsWithResource determines if the path ends with a resource (ID parameter) or collection
func (v *URLPathValidator) endsWithResource(urlPath string) bool {
	segments := strings.Split(urlPath, "/")
	if len(segments) == 0 {
		return false
	}

	lastSegment := segments[len(segments)-1]
	return strings.HasPrefix(lastSegment, "{") && strings.HasSuffix(lastSegment, "}")
}

// determineHTTPMethods determines appropriate HTTP methods based on the path structure
func (v *URLPathValidator) determineHTTPMethods(endsWithResource bool) []string {
	var methods []string

	// GET is valid for both collection and resource paths
	methods = append(methods, "get")

	if endsWithResource {
		// Resource paths (ending with an ID parameter) should use PUT, PATCH, DELETE
		methods = append(methods, "put", "patch", "delete")
	} else {
		// Collection paths should use POST
		methods = append(methods, "post")
	}

	return methods
}

// createOpenAPISpec creates a minimal OpenAPI specification for the URL path
func (v *URLPathValidator) createOpenAPISpec(urlPath string, pathParams []string, methods []string) string {
	// Create path parameters section
	var paramsSection string
	for _, param := range pathParams {
		paramsSection += fmt.Sprintf(`
        - name: %s
          in: path
          required: true
          schema:
            type: string`, param)
	}

	// Create methods section
	var methodsSection string
	for _, method := range methods {
		methodsSection += fmt.Sprintf(`
    %s:
      summary: %s endpoint
      description: Auto-generated %s endpoint for validation
      %s
      responses:
        '200':
          description: OK`,
			method,
			strings.ToUpper(method),
			strings.ToUpper(method),
			func() string {
				if paramsSection != "" && method != "post" {
					return fmt.Sprintf("parameters:%s", paramsSection)
				}
				return ""
			}())
	}

	// Create the OpenAPI specification
	spec := fmt.Sprintf(`openapi: 3.0.0
info:
  title: Auto-generated API for URL Path Validation
  version: 1.0.0
paths:
  %s:%s`,
		urlPath, methodsSection)

	return spec
}
