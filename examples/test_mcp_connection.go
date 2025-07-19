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
	flag.Parse()

	// Test ping
	fmt.Println("Testing ping...")
	if err := testPing(*url); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Ping successful!")

	// Test getTools
	fmt.Println("\nTesting getTools...")
	if err := testGetTools(*url); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("getTools successful!")

	// Test getResources
	fmt.Println("\nTesting getResources...")
	if err := testGetResources(*url); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("getResources successful!")

	fmt.Println("\nAll tests passed!")
}

// testPing tests the ping method
func testPing(url string) error {
	// Create the request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "ping",
		"params":  map[string]interface{}{},
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

	if result != "pong" {
		return fmt.Errorf("unexpected result: %v", result)
	}

	return nil
}

// testGetTools tests the getTools method
func testGetTools(url string) error {
	// Create the request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "getTools",
		"params":  map[string]interface{}{},
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

	// Print the tools
	tools, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected result type: %T", result)
	}

	fmt.Println("Available tools:")
	for name, tool := range tools {
		fmt.Printf("- %s\n", name)
		if toolObj, ok := tool.(map[string]interface{}); ok {
			if desc, ok := toolObj["description"].(string); ok {
				fmt.Printf("  Description: %s\n", desc)
			}
		}
	}

	return nil
}

// testGetResources tests the getResources method
func testGetResources(url string) error {
	// Create the request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "getResources",
		"params":  map[string]interface{}{},
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

	// Print the resources
	resources, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected result type: %T", result)
	}

	fmt.Println("Available resources:")
	for name, resource := range resources {
		fmt.Printf("- %s\n", name)
		if resourceObj, ok := resource.(map[string]interface{}); ok {
			if desc, ok := resourceObj["description"].(string); ok {
				fmt.Printf("  Description: %s\n", desc)
			}
		}
	}

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
