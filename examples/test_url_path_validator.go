package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	// Parse command line flags
	url := flag.String("url", "http://localhost:9090/mcp", "The URL of the MCP server")
	urlPath := flag.String("path", "/api/v0/admin/cloudAgents/{datacenterId}/upgrades", "The URL path to validate")
	flag.Parse()

	// If no URL path provided, use the first argument
	if flag.NArg() > 0 {
		*urlPath = flag.Arg(0)
	}

	// Test URL path validation
	fmt.Printf("Testing URL path validation for: %s\n", *urlPath)
	if err := testURLPathValidation(*url, *urlPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("URL path validation successful!")
}

// testURLPathValidation tests the validateURLPath method
func testURLPathValidation(url, urlPath string) error {
	// Create the request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "validateURLPath",
		"params": map[string]interface{}{
			"url_path": urlPath,
		},
	}

	// Send the request
	response, err := sendRequest(url, request)
	if err != nil {
		return err
	}

	// Check the response
	result, ok := response["result"]
	if !ok {
		return fmt.Errorf("missing result in response")
	}

	// Print the results
	resultsJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("error formatting results: %v", err)
	}
	fmt.Println(string(resultsJSON))

	return nil
}

// sendRequest sends a JSON-RPC request to the MCP server
func sendRequest(url string, request map[string]interface{}) (map[string]interface{}, error) {
	// Convert the request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create an HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	// Check for errors
	if errorObj, ok := response["error"].(map[string]interface{}); ok {
		return nil, fmt.Errorf("error from MCP server: %s", errorObj["message"])
	}

	return response, nil
}
