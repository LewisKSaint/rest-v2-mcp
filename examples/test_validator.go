package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/solacedev/restv2-api-server-go/internal/validator"
)

func main() {
	// Parse command line flags
	apiSpecPath := flag.String("api-spec", "", "Path to the API specification file")
	flag.Parse()

	// If no API spec path provided, use the first argument
	if *apiSpecPath == "" && len(flag.Args()) > 0 {
		*apiSpecPath = flag.Args()[0]
	}

	// Check if API spec path is provided
	if *apiSpecPath == "" {
		fmt.Println("Error: API specification file path is required")
		fmt.Println("Usage: go run test_validator.go [--api-spec] <path-to-api-spec>")
		os.Exit(1)
	}

	// Check if the file exists
	if _, err := os.Stat(*apiSpecPath); os.IsNotExist(err) {
		fmt.Printf("Error: API specification file not found: %s\n", *apiSpecPath)
		os.Exit(1)
	}

	// Read the API spec file
	apiSpecContent, err := ioutil.ReadFile(*apiSpecPath)
	if err != nil {
		fmt.Printf("Error reading API spec file: %v\n", err)
		os.Exit(1)
	}

	// Create a new validator
	v, err := validator.NewValidator()
	if err != nil {
		fmt.Printf("Error creating validator: %v\n", err)
		os.Exit(1)
	}

	// Validate the API spec
	params := map[string]interface{}{
		"api_spec": string(apiSpecContent),
	}

	results, err := v.Validate(params)
	if err != nil {
		fmt.Printf("Error validating API spec: %v\n", err)
		os.Exit(1)
	}

	// Print the results
	fmt.Printf("Validation results for %s:\n\n", filepath.Base(*apiSpecPath))
	resultsJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting results: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(resultsJSON))

	// Check if any rules failed
	if ruleResults, ok := results["results"].(map[string]interface{}); ok {
		failedRules := 0
		for ruleName, ruleResult := range ruleResults {
			if result, ok := ruleResult.(map[string]interface{}); ok {
				if status, ok := result["status"].(string); ok && status == "failed" {
					failedRules++
				}
			}
		}

		if failedRules > 0 {
			fmt.Printf("\n%d rule(s) failed validation.\n", failedRules)
			os.Exit(1)
		} else {
			fmt.Println("\nAll rules passed validation!")
		}
	}
}
