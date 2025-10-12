package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server represents the MCP server
type Server struct {
	port       int
	httpServer *http.Server
	handler    *Handler
}

// NewServer creates a new MCP server instance
func NewServer(port int) *Server {
	handler := NewHandler()

	return &Server{
		port:    port,
		handler: handler,
	}
}

// Start starts the MCP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// MCP protocol endpoints
	mux.HandleFunc("/mcp/v1/initialize", s.handleInitialize)
	mux.HandleFunc("/mcp/v1/tools/list", s.handleToolsList)
	mux.HandleFunc("/mcp/v1/tools/call", s.handleToolsCall)
	mux.HandleFunc("/mcp/v1/resources/list", s.handleResourcesList)
	mux.HandleFunc("/mcp/v1/resources/read", s.handleResourcesRead)
	mux.HandleFunc("/mcp/v1/prompts/list", s.handlePromptsList)
	mux.HandleFunc("/mcp/v1/prompts/get", s.handlePromptsGet)

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go s.handleShutdown()

	log.Printf("MCP Server listening on port %d", s.port)
	return s.httpServer.ListenAndServe()
}

// handleShutdown handles graceful shutdown on interrupt signals
func (s *Server) handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutdown signal received, gracefully stopping server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
	os.Exit(0)
}

// handleInitialize handles the MCP initialize request
func (s *Server) handleInitialize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]string{
			"name":    "technocrat",
			"version": "1.0.0",
		},
		"capabilities": map[string]interface{}{
			"tools":     map[string]bool{},
			"resources": map[string]bool{},
			"prompts":   map[string]bool{},
		},
	}

	s.respondJSON(w, http.StatusOK, response)
}

// handleToolsList handles listing available tools
func (s *Server) handleToolsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := s.handler.ListTools()
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"tools": tools,
	})
}

// handleToolsCall handles tool execution
func (s *Server) handleToolsCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	result, err := s.handler.CallTool(request.Name, request.Arguments)
	if err != nil {
		s.respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"result": result,
	})
}

// handleResourcesList handles listing available resources
func (s *Server) handleResourcesList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resources := s.handler.ListResources()
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"resources": resources,
	})
}

// handleResourcesRead handles reading a resource
func (s *Server) handleResourcesRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request struct {
		URI string `json:"uri"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	content, err := s.handler.ReadResource(request.URI)
	if err != nil {
		s.respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"contents": content,
	})
}

// handlePromptsList handles listing available prompts
func (s *Server) handlePromptsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prompts := s.handler.ListPrompts()
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"prompts": prompts,
	})
}

// handlePromptsGet handles getting a specific prompt
func (s *Server) handlePromptsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	prompt, err := s.handler.GetPrompt(request.Name, request.Arguments)
	if err != nil {
		s.respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	s.respondJSON(w, http.StatusOK, prompt)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// respondJSON sends a JSON response
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}
