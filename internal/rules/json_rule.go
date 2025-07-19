package rules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// JSONRule implements the Rule interface for JSON-defined rules
type JSONRule struct {
	RuleName        string      `json:"name"`
	RuleDescription string      `json:"description"`
	Enabled         bool        `json:"enabled"`
	Conditions      []Condition `json:"conditions"`
	FilePath        string      `json:"-"` // Not part of the JSON, used for reference
}

// Condition represents a validation condition in a JSON rule
type Condition struct {
	Type    string      `json:"type"`
	Pattern string      `json:"pattern,omitempty"`
	Path    string      `json:"path,omitempty"`
	Method  string      `json:"method,omitempty"`
	Field   string      `json:"field,omitempty"`
	Format  string      `json:"format,omitempty"`
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"message"`
}

// NewJSONRuleFromFile creates a new JSONRule from a file
func NewJSONRuleFromFile(filePath string) (*JSONRule, error) {
	// Read the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON rule file: %v", err)
	}

	// Parse the JSON
	var rule JSONRule
	if err := json.Unmarshal(data, &rule); err != nil {
		return nil, fmt.Errorf("error parsing JSON rule: %v", err)
	}

	// Set the file path
	rule.FilePath = filePath

	// Validate the rule
	if err := rule.validate(); err != nil {
		return nil, fmt.Errorf("invalid JSON rule: %v", err)
	}

	return &rule, nil
}

// LoadJSONRulesFromDir loads all JSON rules from a directory
func LoadJSONRulesFromDir(dirPath string) (map[string]Rule, error) {
	rules := make(map[string]Rule)

	// Create the directory if it doesn't exist
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, fmt.Errorf("error creating rules directory: %v", err)
		}
		return rules, nil // Return empty rules since directory was just created
	}

	// Walk the directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non-JSON files
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
			return nil
		}

		// Load the rule
		rule, err := NewJSONRuleFromFile(path)
		if err != nil {
			return fmt.Errorf("error loading rule from %s: %v", path, err)
		}

		// Check for name conflicts
		baseName := rule.RuleName
		suffix := 1
		for {
			if _, exists := rules[rule.RuleName]; !exists {
				break
			}
			rule.RuleName = fmt.Sprintf("%s_%d", baseName, suffix)
			suffix++
		}

		// Add the rule
		rules[rule.RuleName] = rule
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking rules directory: %v", err)
	}

	return rules, nil
}

// validate validates the rule
func (r *JSONRule) validate() error {
	// Check required fields
	if r.RuleName == "" {
		return fmt.Errorf("rule name is required")
	}
	if r.RuleDescription == "" {
		return fmt.Errorf("rule description is required")
	}
	if len(r.Conditions) == 0 {
		return fmt.Errorf("at least one condition is required")
	}

	// Validate conditions
	for i, condition := range r.Conditions {
		if condition.Type == "" {
			return fmt.Errorf("condition %d: type is required", i)
		}
		if condition.Message == "" {
			return fmt.Errorf("condition %d: message is required", i)
		}

		// Validate condition type
		switch condition.Type {
		case "path_pattern":
			if condition.Pattern == "" {
				return fmt.Errorf("condition %d: pattern is required for path_pattern", i)
			}
			// Validate the pattern as a valid regex
			if _, err := regexp.Compile(condition.Pattern); err != nil {
				return fmt.Errorf("condition %d: invalid regex pattern: %v", i, err)
			}
		case "method_check":
			if condition.Path == "" {
				return fmt.Errorf("condition %d: path is required for method_check", i)
			}
			if condition.Method == "" {
				return fmt.Errorf("condition %d: method is required for method_check", i)
			}
		case "parameter_check":
			if condition.Path == "" {
				return fmt.Errorf("condition %d: path is required for parameter_check", i)
			}
		case "resource_naming":
			if condition.Pattern == "" {
				return fmt.Errorf("condition %d: pattern is required for resource_naming", i)
			}
			// Validate the pattern as a valid regex
			if _, err := regexp.Compile(condition.Pattern); err != nil {
				return fmt.Errorf("condition %d: invalid regex pattern: %v", i, err)
			}
		case "schema_field":
			if condition.Field == "" {
				return fmt.Errorf("condition %d: field is required for schema_field", i)
			}
		default:
			return fmt.Errorf("condition %d: unknown type: %s", i, condition.Type)
		}
	}

	return nil
}

// Name returns the name of the rule
func (r *JSONRule) Name() string {
	return r.RuleName
}

// Description returns the description of the rule
func (r *JSONRule) Description() string {
	return r.RuleDescription
}

// Apply applies the rule to the given API spec
func (r *JSONRule) Apply(spec map[string]interface{}) (map[string]interface{}, error) {
	// Skip if the rule is disabled
	if !r.Enabled {
		return map[string]interface{}{
			"status":  "skipped",
			"message": "Rule is disabled",
		}, nil
	}

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

	// Apply each condition
	for _, condition := range r.Conditions {
		switch condition.Type {
		case "path_pattern":
			// Check all paths against the pattern
			pattern := regexp.MustCompile(condition.Pattern)
			for path := range paths {
				if !pattern.MatchString(path) {
					issues = append(issues, map[string]interface{}{
						"path":    path,
						"message": condition.Message,
					})
				}
			}
		case "method_check":
			// Check if the specified path has the specified method
			pathObj, ok := paths[condition.Path].(map[string]interface{})
			if !ok {
				continue
			}
			if _, ok := pathObj[strings.ToLower(condition.Method)]; !ok {
				issues = append(issues, map[string]interface{}{
					"path":    condition.Path,
					"message": condition.Message,
				})
			}
		case "parameter_check":
			// Check if the specified path has parameters
			pathObj, ok := paths[condition.Path].(map[string]interface{})
			if !ok {
				continue
			}
			if _, ok := pathObj["parameters"]; !ok {
				issues = append(issues, map[string]interface{}{
					"path":    condition.Path,
					"message": condition.Message,
				})
			}
		case "resource_naming":
			// Check all paths against the resource naming pattern
			pattern := regexp.MustCompile(condition.Pattern)
			for path := range paths {
				segments := strings.Split(path, "/")
				for _, segment := range segments {
					if segment == "" {
						continue
					}
					if !pattern.MatchString(segment) {
						issues = append(issues, map[string]interface{}{
							"path":    path,
							"segment": segment,
							"message": condition.Message,
						})
						break
					}
				}
			}
		case "schema_field":
			// Check if the specified field is present in the schema definitions
			schemas, ok := spec["components"].(map[string]interface{})
			if !ok {
				// If there are no components, check for definitions (OpenAPI 2.0)
				schemas, ok = spec["definitions"].(map[string]interface{})
				if !ok {
					// No schemas defined, add an issue
					issues = append(issues, map[string]interface{}{
						"field":   condition.Field,
						"message": "No schema definitions found in API spec",
					})
					continue
				}
			} else {
				// For OpenAPI 3.0, schemas are under components.schemas
				schemas, ok = schemas["schemas"].(map[string]interface{})
				if !ok {
					// No schemas defined, add an issue
					issues = append(issues, map[string]interface{}{
						"field":   condition.Field,
						"message": "No schema definitions found in API spec",
					})
					continue
				}
			}

			// Check all schemas for the field
			fieldFound := false
			for schemaName, schemaObj := range schemas {
				schema, ok := schemaObj.(map[string]interface{})
				if !ok {
					continue
				}

				properties, ok := schema["properties"].(map[string]interface{})
				if !ok {
					continue
				}

				if _, ok := properties[condition.Field]; ok {
					// Field found, check format if specified
					if condition.Format != "" {
						fieldObj, ok := properties[condition.Field].(map[string]interface{})
						if !ok {
							continue
						}

						format, ok := fieldObj["format"].(string)
						if !ok || format != condition.Format {
							issues = append(issues, map[string]interface{}{
								"schema":  schemaName,
								"field":   condition.Field,
								"message": fmt.Sprintf("%s (format should be %s)", condition.Message, condition.Format),
							})
						}
					}
					fieldFound = true
					break
				}
			}

			if !fieldFound {
				// Field not found in any schema
				issues = append(issues, map[string]interface{}{
					"field":   condition.Field,
					"message": condition.Message,
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
