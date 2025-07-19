package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/solace/restv2-api-server-go/internal/validator"
)

// Server represents the REST API Validator MCP server
type Server struct {
	port      int
	keepAlive bool
	server    *http.Server
	validator *validator.Validator
	mu        sync.Mutex
	lastPing  time.Time
}

// NewServer creates a new server instance
func NewServer(port int, keepAlive bool) (*Server, error) {
	v, err := validator.NewValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %v", err)
	}

	s := &Server{
		port:      port,
		keepAlive: keepAlive,
		validator: v,
		lastPing:  time.Now(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleMCP)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return s, nil
}

// Start starts the server
func (s *Server) Start() error {
	// Start keep-alive goroutine if enabled
	if s.keepAlive {
		go s.keepAliveMonitor()
	}

	// Start the HTTP server
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// handleMCP handles MCP requests
func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing request: %v", err), http.StatusBadRequest)
		return
	}

	// Check if this is a ping request
	if method, ok := request["method"].(string); ok && method == "ping" {
		s.handlePing(w)
		return
	}

	// Process the request
	response, err := s.processRequest(request)
	if err != nil {
		errorResponse := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      request["id"],
			"error": map[string]interface{}{
				"code":    -32000,
				"message": fmt.Sprintf("Error processing request: %v", err),
			},
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handlePing handles ping requests
func (s *Server) handlePing(w http.ResponseWriter) {
	// Update last ping time
	s.mu.Lock()
	s.lastPing = time.Now()
	s.mu.Unlock()

	// Send pong response
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"result":  "pong",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processRequest processes an MCP request
func (s *Server) processRequest(request map[string]interface{}) (map[string]interface{}, error) {
	// Extract method and params
	method, ok := request["method"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid method")
	}

	params, ok := request["params"].(map[string]interface{})
	if !ok {
		params = make(map[string]interface{})
	}

	// Process the request based on the method
	var result interface{}
	var err error

	switch method {
	case "validate":
		result, err = s.validator.Validate(params)
	case "validateURLPath":
		result, err = s.validator.ValidateURLPath(params)
	case "getTools":
		result = s.getTools()
	case "getResources":
		result = s.getResources()
	default:
		err = fmt.Errorf("unknown method: %s", method)
	}

	if err != nil {
		return nil, err
	}

	// Create the response
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      request["id"],
		"result":  result,
	}

	return response, nil
}

// getTools returns the available tools
func (s *Server) getTools() map[string]interface{} {
	return map[string]interface{}{
		"validate_api": map[string]interface{}{
			"description": "Validate a REST API against a set of rules",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"api_spec": map[string]interface{}{
						"type":        "string",
						"description": "OpenAPI specification in YAML or JSON format",
					},
					"rules": map[string]interface{}{
						"type":        "array",
						"description": "List of rules to validate against",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"api_spec"},
			},
		},
		"validate_url_path": map[string]interface{}{
			"description": "Validate a single URL path against Solace REST API conventions",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url_path": map[string]interface{}{
						"type":        "string",
						"description": "URL path to validate (e.g., /api/v0/admin/cloudAgents/{datacenterId}/upgrades)",
					},
				},
				"required": []string{"url_path"},
			},
		},
	}
}

// getResources returns the available resources
func (s *Server) getResources() map[string]interface{} {
	return map[string]interface{}{
		"rules": map[string]interface{}{
			"description": "Available validation rules",
		},
	}
}

// keepAliveMonitor monitors the server for inactivity
func (s *Server) keepAliveMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		s.mu.Lock()
		elapsed := time.Since(s.lastPing)
		s.mu.Unlock()

		// If no ping received in the last 5 minutes, send a ping to self
		if elapsed > 5*time.Minute {
			log.Println("No activity detected for 5 minutes, sending self-ping...")
			s.sendSelfPing()
		}
	}
}

// sendSelfPing sends a ping request to the server itself
func (s *Server) sendSelfPing() {
	// Create a ping request
	pingRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "ping",
		"params":  map[string]interface{}{},
	}

	// Convert to JSON
	requestJSON, err := json.Marshal(pingRequest)
	if err != nil {
		log.Printf("Error creating ping request: %v", err)
		return
	}

	// Send the request
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/mcp", s.port), "application/json", nil)
	if err != nil {
		log.Printf("Error sending ping request: %v : %v", err, requestJSON)
		return
	}
	defer resp.Body.Close()

	log.Println("Self-ping sent successfully")
}
