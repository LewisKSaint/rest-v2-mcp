package rules

import (
	"fmt"
	"strings"
)

// Rule defines the interface for a validation rule
type Rule interface {
	Apply(spec map[string]interface{}) (map[string]interface{}, error)
	Name() string
	Description() string
}

// SolaceRestRules implements the Solace REST API rules
type SolaceRestRules struct{}

// NewSolaceRestRules creates a new SolaceRestRules instance
func NewSolaceRestRules() *SolaceRestRules {
	return &SolaceRestRules{}
}

// Apply applies the Solace REST API rules to the given API spec
func (r *SolaceRestRules) Apply(spec map[string]interface{}) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	issues := []map[string]interface{}{}

	// Check if the spec is valid
	if spec == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "invalid API spec",
		}, nil
	}

	// Check if the spec has paths
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{
			"status":  "error",
			"message": "API spec does not have paths",
		}, nil
	}

	// Check if the paths follow REST conventions
	for path, pathObj := range paths {
		pathItem, ok := pathObj.(map[string]interface{})
		if !ok {
			continue
		}

		// Check HTTP methods
		for method, methodObj := range pathItem {
			if method == "parameters" || method == "summary" || method == "description" {
				continue
			}

			methodDef, ok := methodObj.(map[string]interface{})
			if !ok {
				continue
			}

			// Check if the method is appropriate for the path
			if issue := r.checkMethodPathConsistency(path, method, methodDef); issue != nil {
				issues = append(issues, issue)
			}
		}
	}

	// Set the results
	if len(issues) > 0 {
		results["status"] = "failed"
		results["issues"] = issues
	} else {
		results["status"] = "passed"
	}

	return results, nil
}

// checkMethodPathConsistency checks if the HTTP method is appropriate for the path
func (r *SolaceRestRules) checkMethodPathConsistency(path, method string, methodDef map[string]interface{}) map[string]interface{} {
	// Convert method to lowercase for comparison
	method = strings.ToLower(method)

	// Check if the path ends with an ID parameter
	endsWithID := strings.Contains(path, "/{") && strings.HasSuffix(path, "}")

	// Rules for REST API methods
	switch method {
	case "get":
		// GET is valid for both collection and resource paths
		return nil
	case "post":
		// POST should be used for collection paths
		if endsWithID {
			return map[string]interface{}{
				"path":    path,
				"method":  method,
				"message": "POST should be used for collection paths, not for specific resources",
			}
		}
	case "put", "patch", "delete":
		// PUT, PATCH, DELETE should be used for resource paths
		if !endsWithID {
			return map[string]interface{}{
				"path":    path,
				"method":  method,
				"message": fmt.Sprintf("%s should be used for specific resources, not for collections", strings.ToUpper(method)),
			}
		}
	}

	return nil
}

// Name returns the name of the rule
func (r *SolaceRestRules) Name() string {
	return "solace_rest_rules"
}

// Description returns the description of the rule
func (r *SolaceRestRules) Description() string {
	return "Validates that the API follows Solace REST API conventions"
}

// SolaceSingularUserResourcesRule implements the Solace singular user resources rule
type SolaceSingularUserResourcesRule struct{}

// NewSolaceSingularUserResourcesRule creates a new SolaceSingularUserResourcesRule instance
func NewSolaceSingularUserResourcesRule() *SolaceSingularUserResourcesRule {
	return &SolaceSingularUserResourcesRule{}
}

// Apply applies the Solace singular user resources rule to the given API spec
func (r *SolaceSingularUserResourcesRule) Apply(spec map[string]interface{}) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	issues := []map[string]interface{}{}

	// Check if the spec is valid
	if spec == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "invalid API spec",
		}, nil
	}

	// Check if the spec has paths
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{
			"status":  "error",
			"message": "API spec does not have paths",
		}, nil
	}

	// Check for user-specific resources
	for path := range paths {
		if strings.Contains(path, "/users/") && !strings.Contains(path, "/me/") {
			// This is a user-specific resource that doesn't use /me/
			issues = append(issues, map[string]interface{}{
				"path":    path,
				"message": "User-specific resources should use /me/ instead of /users/{id}",
			})
		}
	}

	// Set the results
	if len(issues) > 0 {
		results["status"] = "failed"
		results["issues"] = issues
	} else {
		results["status"] = "passed"
	}

	return results, nil
}

// Name returns the name of the rule
func (r *SolaceSingularUserResourcesRule) Name() string {
	return "solace_singular_user_resources"
}

// Description returns the description of the rule
func (r *SolaceSingularUserResourcesRule) Description() string {
	return "Validates that user-specific resources use /me/ instead of /users/{id}"
}

// SolaceCustomActionsRule implements the Solace custom actions rule
type SolaceCustomActionsRule struct{}

// NewSolaceCustomActionsRule creates a new SolaceCustomActionsRule instance
func NewSolaceCustomActionsRule() *SolaceCustomActionsRule {
	return &SolaceCustomActionsRule{}
}

// Apply applies the Solace custom actions rule to the given API spec
func (r *SolaceCustomActionsRule) Apply(spec map[string]interface{}) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	issues := []map[string]interface{}{}

	// Check if the spec is valid
	if spec == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "invalid API spec",
		}, nil
	}

	// Check if the spec has paths
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{
			"status":  "error",
			"message": "API spec does not have paths",
		}, nil
	}

	// Check for custom actions
	for path, pathObj := range paths {
		pathItem, ok := pathObj.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if the path contains a custom action
		if strings.Contains(path, "/actions/") {
			// Check if the action is properly defined
			if _, ok := pathItem["post"]; !ok {
				issues = append(issues, map[string]interface{}{
					"path":    path,
					"message": "Custom actions should use POST method",
				})
			}
		}
	}

	// Set the results
	if len(issues) > 0 {
		results["status"] = "failed"
		results["issues"] = issues
	} else {
		results["status"] = "passed"
	}

	return results, nil
}

// Name returns the name of the rule
func (r *SolaceCustomActionsRule) Name() string {
	return "solace_custom_actions"
}

// Description returns the description of the rule
func (r *SolaceCustomActionsRule) Description() string {
	return "Validates that custom actions follow Solace conventions"
}
