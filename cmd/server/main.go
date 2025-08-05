package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/solacedev/restv2-api-server-go/internal/server"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 9090, "The port to listen on")
	keepAlive := flag.Bool("keep-alive", false, "Enable keep-alive mechanism")
	flag.Parse()

	// Check environment variables
	if envPort := os.Getenv("MCP_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			*port = p
		}
	}

	if os.Getenv("MCP_KEEP_ALIVE") == "true" {
		*keepAlive = true
	}

	// Create a new server
	s, err := server.NewServer(*port, *keepAlive)
	if err != nil {
		log.Fatalf("Error creating server: %v", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	errChan := make(chan error)
	go func() {
		fmt.Printf("Starting REST API Validator MCP server on port %d...\n", *port)
		if *keepAlive {
			fmt.Println("Keep-alive mechanism enabled")
		}
		errChan <- s.Start()
	}()

	// Wait for termination signal or server error
	select {
	case sig := <-sigChan:
		fmt.Printf("Received signal %v, shutting down...\n", sig)
	case err := <-errChan:
		if err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}

	// Shutdown the server
	if err := s.Shutdown(); err != nil {
		fmt.Printf("Error shutting down server: %v\n", err)
	}
}
