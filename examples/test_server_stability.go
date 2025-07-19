package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Parse command line flags
	url := flag.String("url", "http://localhost:9090/mcp", "The URL of the MCP server")
	interval := flag.Int("interval", 30, "The interval between pings in seconds")
	duration := flag.Int("duration", 3600, "The total duration of the test in seconds")
	maxConsecutiveFailures := flag.Int("max-failures", 3, "Maximum consecutive failures before attempting reconnection")
	flag.Parse()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the stability test in a goroutine
	doneChan := make(chan struct{})
	go func() {
		testServerStability(*url, *interval, *duration, *maxConsecutiveFailures)
		close(doneChan)
	}()

	// Wait for termination signal or test completion
	select {
	case sig := <-sigChan:
		fmt.Printf("Received signal %v, shutting down...\n", sig)
	case <-doneChan:
		fmt.Println("Stability test completed")
	}
}

func testServerStability(url string, interval, duration, maxConsecutiveFailures int) {
	startTime := time.Now()
	endTime := startTime.Add(time.Duration(duration) * time.Second)
	successCount := 0
	failureCount := 0
	consecutiveFailures := 0

	fmt.Printf("Starting stability test for %s\n", url)
	fmt.Printf("Test will run for %d seconds with %d second intervals\n", duration, interval)
	fmt.Printf("Will attempt reconnection after %d consecutive failures\n", maxConsecutiveFailures)

	for time.Now().Before(endTime) {
		// Try to ping the server
		err := pingServer(url)
		if err == nil {
			// Log success
			successCount++
			consecutiveFailures = 0 // Reset consecutive failures counter
			fmt.Printf("[%s] Ping successful (%d successes, %d failures)\n",
				time.Now().Format("2006-01-02 15:04:05"), successCount, failureCount)
		} else {
			// Log failure
			failureCount++
			consecutiveFailures++
			fmt.Printf("[%s] Ping failed: %v\n",
				time.Now().Format("2006-01-02 15:04:05"), err)
			fmt.Printf("Total failures: %d, Consecutive failures: %d\n", failureCount, consecutiveFailures)

			// If we've had too many consecutive failures, wait a bit longer before retrying
			if consecutiveFailures >= maxConsecutiveFailures {
				fmt.Printf("Reached %d consecutive failures. Waiting for 30 seconds before retrying...\n", consecutiveFailures)
				time.Sleep(30 * time.Second)
				fmt.Println("Attempting to reconnect...")
				consecutiveFailures = 0 // Reset after waiting
				continue
			}
		}

		// Wait for the next interval
		time.Sleep(time.Duration(interval) * time.Second)
	}

	// Log final results
	totalPings := successCount + failureCount
	successRate := float64(successCount) / float64(totalPings) * 100
	if totalPings == 0 {
		successRate = 0
	}

	fmt.Println("Stability test completed")
	fmt.Printf("Total pings: %d\n", totalPings)
	fmt.Printf("Successful pings: %d\n", successCount)
	fmt.Printf("Failed pings: %d\n", failureCount)
	fmt.Printf("Success rate: %.2f%%\n", successRate)
}

func pingServer(url string) error {
	// Create the request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "ping",
		"params":  map[string]interface{}{},
	}

	// Convert the request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshaling request: %v", err)
	}

	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Send the request
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("error parsing response: %v", err)
	}

	// Check for errors
	if errorObj, ok := response["error"].(map[string]interface{}); ok {
		return fmt.Errorf("error from MCP server: %s", errorObj["message"])
	}

	// Check the result
	result, ok := response["result"]
	if !ok {
		return fmt.Errorf("missing result in response")
	}

	// Check if the result is "pong"
	if result != "pong" {
		return fmt.Errorf("unexpected result: %v", result)
	}

	return nil
}
