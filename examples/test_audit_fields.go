package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/solace/restv2-api-server-go/internal/validator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run examples/test_audit_fields.go <api-spec-file>")
		os.Exit(1)
	}

	apiSpecFile := os.Args[1]
	fmt.Printf("Testing audit fields validation for: %s\n", apiSpecFile)

	// Create a validator
	v, err := validator.NewValidator()
	if err != nil {
		fmt.Printf("Error creating validator: %v\n", err)
		os.Exit(1)
	}

	// Validate the API spec with specific focus on audit fields
	results, err := v.Validate(map[string]interface{}{
		"api_spec": apiSpecFile,
		"rules":    []interface{}{"audit_fields"},
	})
	if err != nil {
		fmt.Printf("Error validating API spec: %v\n", err)
		os.Exit(1)
	}

	// Print the results
	resultsJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling results: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(resultsJSON))

	// Check if validation passed
	auditFieldsResult, ok := results["results"].(map[string]interface{})["audit_fields"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: audit_fields result not found")
		os.Exit(1)
	}

	if auditFieldsResult["status"] == "passed" {
		fmt.Println("Audit fields validation successful!")
	} else {
		fmt.Println("Audit fields validation failed!")
		os.Exit(1)
	}
}
