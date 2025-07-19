package validator

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/solace/restv2-api-server-go/internal/rules"
	"gopkg.in/yaml.v3"
)

// Validator represents the REST API validator
type Validator struct {
	rules            map[string]rules.Rule
	urlPathValidator *URLPathValidator
}

// NewValidator creates a new validator instance
func NewValidator() (*Validator, error) {
	v := &Validator{
		rules: make(map[string]rules.Rule),
	}

	// Register built-in rules
	v.registerBuiltInRules()

	// Register JSON rules
	if err := v.registerJSONRules("config/rules"); err != nil {
		// Log the error but continue
		fmt.Printf("Warning: Error loading JSON rules: %v\n", err)
	}

	// Initialize URL path validator
	v.urlPathValidator = NewURLPathValidator(v)

	return v, nil
}

// registerJSONRules registers rules from JSON files in the specified directory
func (v *Validator) registerJSONRules(dirPath string) error {
	// Load JSON rules
	jsonRules, err := rules.LoadJSONRulesFromDir(dirPath)
	if err != nil {
		return fmt.Errorf("error loading JSON rules: %v", err)
	}

	// Register the rules
	for name, rule := range jsonRules {
		v.rules[name] = rule
	}

	return nil
}

// registerBuiltInRules registers the built-in validation rules
func (v *Validator) registerBuiltInRules() {
	// Register Solace REST rules
	v.rules["solace_rest_rules"] = rules.NewSolaceRestRules()
	v.rules["solace_singular_user_resources"] = rules.NewSolaceSingularUserResourcesRule()
	v.rules["solace_custom_actions"] = rules.NewSolaceCustomActionsRule()
}

// Validate validates an API specification against a set of rules
func (v *Validator) Validate(params map[string]interface{}) (map[string]interface{}, error) {
	// Extract API spec from params
	apiSpec, ok := params["api_spec"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid api_spec parameter")
	}

	// Parse the API spec
	spec, err := v.parseAPISpec(apiSpec)
	if err != nil {
		return nil, fmt.Errorf("error parsing API spec: %v", err)
	}

	// Extract rules to validate against
	var rulesToApply []string
	if rulesParam, ok := params["rules"].([]interface{}); ok {
		for _, r := range rulesParam {
			if ruleName, ok := r.(string); ok {
				rulesToApply = append(rulesToApply, ruleName)
			}
		}
	}

	// If no rules specified, use all available rules
	if len(rulesToApply) == 0 {
		for ruleName := range v.rules {
			rulesToApply = append(rulesToApply, ruleName)
		}
	}

	// Apply the rules
	results := make(map[string]interface{})
	for _, ruleName := range rulesToApply {
		rule, ok := v.rules[ruleName]
		if !ok {
			results[ruleName] = map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("rule not found: %s", ruleName),
			}
			continue
		}

		// Apply the rule
		ruleResult, err := rule.Apply(spec)
		if err != nil {
			results[ruleName] = map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("error applying rule: %v", err),
			}
			continue
		}

		results[ruleName] = ruleResult
	}

	return map[string]interface{}{
		"results": results,
	}, nil
}

// parseAPISpec parses an API specification from a string or file path
func (v *Validator) parseAPISpec(apiSpec string) (map[string]interface{}, error) {
	var specContent []byte
	var err error

	// Check if apiSpec is a file path
	if _, err := os.Stat(apiSpec); err == nil {
		// Read the file
		specContent, err = ioutil.ReadFile(apiSpec)
		if err != nil {
			return nil, fmt.Errorf("error reading API spec file: %v", err)
		}
	} else {
		// Treat as raw content
		specContent = []byte(apiSpec)
	}

	// Determine if it's JSON or YAML based on the file extension or content
	var spec map[string]interface{}
	if strings.HasSuffix(apiSpec, ".json") || (len(specContent) > 0 && specContent[0] == '{') {
		// Parse as JSON
		err = yaml.Unmarshal(specContent, &spec)
	} else {
		// Parse as YAML
		err = yaml.Unmarshal(specContent, &spec)
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing API spec: %v", err)
	}

	return spec, nil
}

// GetRules returns the available rules
func (v *Validator) GetRules() []string {
	var ruleNames []string
	for name := range v.rules {
		ruleNames = append(ruleNames, name)
	}
	return ruleNames
}

// ValidateURLPath validates a single URL path against Solace REST API conventions
func (v *Validator) ValidateURLPath(params map[string]interface{}) (map[string]interface{}, error) {
	// Extract URL path from params
	urlPath, ok := params["url_path"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid url_path parameter")
	}

	// Validate the URL path
	return v.urlPathValidator.ValidateURLPath(urlPath)
}
