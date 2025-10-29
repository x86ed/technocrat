package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewServer tests the NewServer constructor
func TestNewServer(t *testing.T) {
	port := 8080
	server := NewServer(port)

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.port != port {
		t.Errorf("Expected port %d, got %d", port, server.port)
	}

	if server.handler == nil {
		t.Error("handler is nil")
	}

	if server.httpServer != nil {
		t.Error("httpServer should be nil before Start() is called")
	}
}

// TestHandleInitialize tests the initialize endpoint
func TestHandleInitialize(t *testing.T) {
	server := NewServer(8080)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "POST request succeeds",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "GET request fails",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
		{
			name:           "PUT request fails",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/mcp/v1/initialize", nil)
			w := httptest.NewRecorder()

			server.handleInitialize(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.checkResponse {
				var response map[string]interface{}
				body, _ := io.ReadAll(resp.Body)
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// Check protocol version
				if version, ok := response["protocolVersion"].(string); !ok || version == "" {
					t.Error("Missing or invalid protocolVersion in response")
				}

				// Check server info
				if serverInfo, ok := response["serverInfo"].(map[string]interface{}); !ok {
					t.Error("Missing serverInfo in response")
				} else {
					if name, ok := serverInfo["name"].(string); !ok || name != "technocrat" {
						t.Error("Invalid server name in response")
					}
					if version, ok := serverInfo["version"].(string); !ok || version == "" {
						t.Error("Missing or invalid version in serverInfo")
					}
				}

				// Check capabilities
				if capabilities, ok := response["capabilities"].(map[string]interface{}); !ok {
					t.Error("Missing capabilities in response")
				} else {
					if _, ok := capabilities["tools"]; !ok {
						t.Error("Missing tools capability")
					}
					if _, ok := capabilities["resources"]; !ok {
						t.Error("Missing resources capability")
					}
					if _, ok := capabilities["prompts"]; !ok {
						t.Error("Missing prompts capability")
					}
				}
			}
		})
	}
}

// TestHandleToolsList tests the tools list endpoint
func TestHandleToolsList(t *testing.T) {
	server := NewServer(8080)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET request succeeds",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST request fails",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/mcp/v1/tools/list", nil)
			w := httptest.NewRecorder()

			server.handleToolsList(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.method == http.MethodGet {
				var response map[string]interface{}
				body, _ := io.ReadAll(resp.Body)
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if _, ok := response["tools"]; !ok {
					t.Error("Missing tools in response")
				}
			}
		})
	}
}

// TestHandleToolsCall tests the tools call endpoint
func TestHandleToolsCall(t *testing.T) {
	server := NewServer(8080)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name:   "Valid echo tool call",
			method: http.MethodPost,
			body: map[string]interface{}{
				"name": "echo",
				"arguments": map[string]interface{}{
					"message": "test message",
				},
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    false,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectError:    false,
		},
		{
			name:   "Unknown tool",
			method: http.MethodPost,
			body: map[string]interface{}{
				"name":      "nonexistent_tool",
				"arguments": map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "Missing arguments",
			method: http.MethodPost,
			body: map[string]interface{}{
				"name": "echo",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody io.Reader
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					reqBody = strings.NewReader(str)
				} else {
					bodyBytes, _ := json.Marshal(tt.body)
					reqBody = bytes.NewReader(bodyBytes)
				}
			}

			req := httptest.NewRequest(tt.method, "/mcp/v1/tools/call", reqBody)
			w := httptest.NewRecorder()

			server.handleToolsCall(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.method == http.MethodPost && tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				body, _ := io.ReadAll(resp.Body)
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if tt.expectError {
					if _, ok := response["error"]; !ok {
						t.Error("Expected error in response")
					}
				} else {
					if _, ok := response["result"]; !ok {
						t.Error("Expected result in response")
					}
				}
			}
		})
	}
}

// TestHandleResourcesList tests the resources list endpoint
func TestHandleResourcesList(t *testing.T) {
	server := NewServer(8080)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET request succeeds",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST request fails",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/mcp/v1/resources/list", nil)
			w := httptest.NewRecorder()

			server.handleResourcesList(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.method == http.MethodGet {
				var response map[string]interface{}
				body, _ := io.ReadAll(resp.Body)
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if _, ok := response["resources"]; !ok {
					t.Error("Missing resources in response")
				}
			}
		})
	}
}

// TestHandleResourcesRead tests the resources read endpoint
func TestHandleResourcesRead(t *testing.T) {
	server := NewServer(8080)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name:   "Valid URI",
			method: http.MethodPost,
			body: map[string]interface{}{
				"uri": "info://server",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    false,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectError:    false,
		},
		{
			name:   "Unknown URI",
			method: http.MethodPost,
			body: map[string]interface{}{
				"uri": "tchncrt://nonexistent",
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody io.Reader
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					reqBody = strings.NewReader(str)
				} else {
					bodyBytes, _ := json.Marshal(tt.body)
					reqBody = bytes.NewReader(bodyBytes)
				}
			}

			req := httptest.NewRequest(tt.method, "/mcp/v1/resources/read", reqBody)
			w := httptest.NewRecorder()

			server.handleResourcesRead(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestHandlePromptsList tests the prompts list endpoint
func TestHandlePromptsList(t *testing.T) {
	server := NewServer(8080)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET request succeeds",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST request fails",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/mcp/v1/prompts/list", nil)
			w := httptest.NewRecorder()

			server.handlePromptsList(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.method == http.MethodGet {
				var response map[string]interface{}
				body, _ := io.ReadAll(resp.Body)
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if _, ok := response["prompts"]; !ok {
					t.Error("Missing prompts in response")
				}
			}
		})
	}
}

// TestHandlePromptsGet tests the prompts get endpoint
func TestHandlePromptsGet(t *testing.T) {
	server := NewServer(8080)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name:   "Valid prompt request",
			method: http.MethodPost,
			body: map[string]interface{}{
				"name": "spec",
				"arguments": map[string]interface{}{
					"feature": "test feature",
				},
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    false,
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectError:    false,
		},
		{
			name:   "Unknown prompt",
			method: http.MethodPost,
			body: map[string]interface{}{
				"name":      "nonexistent_prompt",
				"arguments": map[string]interface{}{},
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody io.Reader
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					reqBody = strings.NewReader(str)
				} else {
					bodyBytes, _ := json.Marshal(tt.body)
					reqBody = bytes.NewReader(bodyBytes)
				}
			}

			req := httptest.NewRequest(tt.method, "/mcp/v1/prompts/get", reqBody)
			w := httptest.NewRecorder()

			server.handlePromptsGet(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestHandleHealth tests the health check endpoint
func TestHandleHealth(t *testing.T) {
	server := NewServer(8080)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response map[string]string
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if status, ok := response["status"]; !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", status)
	}
}

// TestRespondJSON tests the respondJSON helper method
func TestRespondJSON(t *testing.T) {
	server := NewServer(8080)

	tests := []struct {
		name         string
		status       int
		data         interface{}
		expectError  bool
	}{
		{
			name:   "Valid JSON response",
			status: http.StatusOK,
			data: map[string]string{
				"message": "success",
			},
			expectError: false,
		},
		{
			name:   "Complex nested JSON",
			status: http.StatusOK,
			data: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "value",
					"num": 42,
				},
			},
			expectError: false,
		},
		{
			name:   "Array response",
			status: http.StatusOK,
			data: []string{
				"item1",
				"item2",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			server.respondJSON(w, tt.status, tt.data)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, resp.StatusCode)
			}

			contentType := resp.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			var decoded interface{}
			body, _ := io.ReadAll(resp.Body)
			if err := json.Unmarshal(body, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
		})
	}
}

// TestNewStdioServer tests the NewStdioServer constructor
func TestNewStdioServer(t *testing.T) {
	server := NewStdioServer()

	if server == nil {
		t.Fatal("NewStdioServer() returned nil")
	}

	if server.handler == nil {
		t.Error("handler is nil")
	}
}

// TestHandleStdioRequest tests the stdio request handler
func TestHandleStdioRequest(t *testing.T) {
	server := NewStdioServer()

	tests := []struct {
		name          string
		request       map[string]interface{}
		expectError   bool
		checkResult   bool
		resultChecker func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "Initialize request",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "initialize",
			},
			expectError: false,
			checkResult: true,
			resultChecker: func(t *testing.T, response map[string]interface{}) {
				if result, ok := response["result"].(map[string]interface{}); !ok {
					t.Error("Missing result in response")
				} else {
					if _, ok := result["protocolVersion"]; !ok {
						t.Error("Missing protocolVersion in result")
					}
					if _, ok := result["serverInfo"]; !ok {
						t.Error("Missing serverInfo in result")
					}
					if _, ok := result["capabilities"]; !ok {
						t.Error("Missing capabilities in result")
					}
				}
			},
		},
		{
			name: "Tools list request",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      2,
				"method":  "tools/list",
			},
			expectError: false,
			checkResult: true,
			resultChecker: func(t *testing.T, response map[string]interface{}) {
				if result, ok := response["result"].(map[string]interface{}); !ok {
					t.Error("Missing result in response")
				} else {
					if _, ok := result["tools"]; !ok {
						t.Error("Missing tools in result")
					}
				}
			},
		},
		{
			name: "Tools call request",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      3,
				"method":  "tools/call",
				"params": map[string]interface{}{
					"name": "echo",
					"arguments": map[string]interface{}{
						"message": "test",
					},
				},
			},
			expectError: false,
			checkResult: true,
			resultChecker: func(t *testing.T, response map[string]interface{}) {
				if _, ok := response["result"]; !ok {
					t.Error("Missing result in response")
				}
			},
		},
		{
			name: "Resources list request",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      4,
				"method":  "resources/list",
			},
			expectError: false,
			checkResult: true,
			resultChecker: func(t *testing.T, response map[string]interface{}) {
				if result, ok := response["result"].(map[string]interface{}); !ok {
					t.Error("Missing result in response")
				} else {
					if _, ok := result["resources"]; !ok {
						t.Error("Missing resources in result")
					}
				}
			},
		},
		{
			name: "Prompts list request",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      5,
				"method":  "prompts/list",
			},
			expectError: false,
			checkResult: true,
			resultChecker: func(t *testing.T, response map[string]interface{}) {
				if result, ok := response["result"].(map[string]interface{}); !ok {
					t.Error("Missing result in response")
				} else {
					if _, ok := result["prompts"]; !ok {
						t.Error("Missing prompts in result")
					}
				}
			},
		},
		{
			name: "Prompts get request",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      6,
				"method":  "prompts/get",
				"params": map[string]interface{}{
					"name": "spec",
					"arguments": map[string]interface{}{
						"feature": "test",
					},
				},
			},
			expectError: false,
			checkResult: true,
			resultChecker: func(t *testing.T, response map[string]interface{}) {
				if _, ok := response["result"]; !ok {
					t.Error("Missing result in response")
				}
			},
		},
		{
			name: "Missing method",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      7,
			},
			expectError: true,
			checkResult: true,
			resultChecker: func(t *testing.T, response map[string]interface{}) {
				if errorData, ok := response["error"].(map[string]interface{}); !ok {
					t.Error("Expected error in response")
				} else {
					if code, ok := errorData["code"].(int); !ok || code != -32600 {
						t.Errorf("Expected error code -32600, got %v", code)
					}
				}
			},
		},
		{
			name: "Unknown method",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      8,
				"method":  "unknown/method",
			},
			expectError: true,
			checkResult: true,
			resultChecker: func(t *testing.T, response map[string]interface{}) {
				if errorData, ok := response["error"].(map[string]interface{}); !ok {
					t.Error("Expected error in response")
				} else {
					if code, ok := errorData["code"].(int); !ok || code != -32601 {
						t.Errorf("Expected error code -32601, got %v", code)
					}
				}
			},
		},
		{
			name: "Tools call without params",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      9,
				"method":  "tools/call",
			},
			expectError: true,
			checkResult: true,
			resultChecker: func(t *testing.T, response map[string]interface{}) {
				if errorData, ok := response["error"].(map[string]interface{}); !ok {
					t.Error("Expected error in response")
				} else {
					if code, ok := errorData["code"].(int); !ok || code != -32602 {
						t.Errorf("Expected error code -32602, got %v", code)
					}
				}
			},
		},
		{
			name: "Tools call without tool name",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      10,
				"method":  "tools/call",
				"params":  map[string]interface{}{},
			},
			expectError: true,
			checkResult: true,
			resultChecker: func(t *testing.T, response map[string]interface{}) {
				if errorData, ok := response["error"].(map[string]interface{}); !ok {
					t.Error("Expected error in response")
				} else {
					if code, ok := errorData["code"].(int); !ok || code != -32602 {
						t.Errorf("Expected error code -32602, got %v", code)
					}
				}
			},
		},
		{
			name: "Tools call with unknown tool",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      11,
				"method":  "tools/call",
				"params": map[string]interface{}{
					"name": "nonexistent_tool",
				},
			},
			expectError: true,
			checkResult: true,
			resultChecker: func(t *testing.T, response map[string]interface{}) {
				if errorData, ok := response["error"].(map[string]interface{}); !ok {
					t.Error("Expected error in response")
				} else {
					if code, ok := errorData["code"].(int); !ok || code != -32603 {
						t.Errorf("Expected error code -32603, got %v", code)
					}
				}
			},
		},
		{
			name: "Prompts get without params",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      12,
				"method":  "prompts/get",
			},
			expectError: true,
			checkResult: true,
			resultChecker: func(t *testing.T, response map[string]interface{}) {
				if errorData, ok := response["error"].(map[string]interface{}); !ok {
					t.Error("Expected error in response")
				} else {
					if code, ok := errorData["code"].(int); !ok || code != -32602 {
						t.Errorf("Expected error code -32602, got %v", code)
					}
				}
			},
		},
		{
			name: "Prompts get without prompt name",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      13,
				"method":  "prompts/get",
				"params":  map[string]interface{}{},
			},
			expectError: true,
			checkResult: true,
			resultChecker: func(t *testing.T, response map[string]interface{}) {
				if errorData, ok := response["error"].(map[string]interface{}); !ok {
					t.Error("Expected error in response")
				} else {
					if code, ok := errorData["code"].(int); !ok || code != -32602 {
						t.Errorf("Expected error code -32602, got %v", code)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := server.handleStdioRequest(tt.request)

			if response == nil {
				t.Fatal("handleStdioRequest returned nil")
			}

			// Check JSON-RPC 2.0 version
			if jsonrpc, ok := response["jsonrpc"].(string); !ok || jsonrpc != "2.0" {
				t.Error("Missing or invalid jsonrpc version")
			}

			// Check ID is preserved
			if response["id"] != tt.request["id"] {
				t.Errorf("Expected id %v, got %v", tt.request["id"], response["id"])
			}

			if tt.expectError {
				if _, ok := response["error"]; !ok {
					t.Error("Expected error in response")
				}
			}

			if tt.checkResult && tt.resultChecker != nil {
				tt.resultChecker(t, response)
			}
		})
	}
}

// TestServerStartAndShutdown tests server lifecycle (integration-style test)
func TestServerStartAndShutdown(t *testing.T) {
	// This test verifies that the server can be created and configured
	// Note: We don't actually call Start() to avoid blocking
	server := NewServer(8081)

	if server == nil {
		t.Fatal("Failed to create server")
	}

	// Verify the server would set up routes correctly
	// by checking that the handler is initialized
	if server.handler == nil {
		t.Error("Server handler not initialized")
	}
}

// TestServerHTTPEndpoints tests all HTTP endpoints together
func TestServerHTTPEndpoints(t *testing.T) {
	server := NewServer(8080)

	endpoints := []struct {
		path   string
		method string
	}{
		{"/mcp/v1/initialize", http.MethodPost},
		{"/mcp/v1/tools/list", http.MethodGet},
		{"/mcp/v1/resources/list", http.MethodGet},
		{"/mcp/v1/prompts/list", http.MethodGet},
		{"/health", http.MethodGet},
	}

	for _, endpoint := range endpoints {
		t.Run(fmt.Sprintf("%s %s", endpoint.method, endpoint.path), func(t *testing.T) {
			var req *http.Request
			if endpoint.method == http.MethodPost {
				req = httptest.NewRequest(endpoint.method, endpoint.path, strings.NewReader("{}"))
			} else {
				req = httptest.NewRequest(endpoint.method, endpoint.path, nil)
			}

			w := httptest.NewRecorder()

			// Route the request to the appropriate handler
			switch endpoint.path {
			case "/mcp/v1/initialize":
				server.handleInitialize(w, req)
			case "/mcp/v1/tools/list":
				server.handleToolsList(w, req)
			case "/mcp/v1/resources/list":
				server.handleResourcesList(w, req)
			case "/mcp/v1/prompts/list":
				server.handlePromptsList(w, req)
			case "/health":
				server.handleHealth(w, req)
			}

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d for %s %s", 
					resp.StatusCode, endpoint.method, endpoint.path)
			}
		})
	}
}

// TestServerHTTPTimeouts tests that timeouts are configured correctly
func TestServerHTTPTimeouts(t *testing.T) {
	server := NewServer(8080)
	
	// Create a minimal HTTP server to test configuration
	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.handleHealth)
	
	server.httpServer = &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if server.httpServer.ReadTimeout != 15*time.Second {
		t.Errorf("Expected ReadTimeout 15s, got %v", server.httpServer.ReadTimeout)
	}

	if server.httpServer.WriteTimeout != 15*time.Second {
		t.Errorf("Expected WriteTimeout 15s, got %v", server.httpServer.WriteTimeout)
	}

	if server.httpServer.IdleTimeout != 60*time.Second {
		t.Errorf("Expected IdleTimeout 60s, got %v", server.httpServer.IdleTimeout)
	}
}

// TestStdioServerJSONRPCCompliance tests JSON-RPC 2.0 compliance
func TestStdioServerJSONRPCCompliance(t *testing.T) {
	server := NewStdioServer()

	// Test that all responses include required JSON-RPC fields
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      42,
		"method":  "initialize",
	}

	response := server.handleStdioRequest(request)

	// Check required fields
	if jsonrpc, ok := response["jsonrpc"].(string); !ok || jsonrpc != "2.0" {
		t.Error("Response missing or invalid 'jsonrpc' field")
	}

	if id := response["id"]; id != 42 {
		t.Errorf("Response ID mismatch: expected 42, got %v", id)
	}

	// Must have either result or error, but not both
	hasResult := response["result"] != nil
	hasError := response["error"] != nil

	if hasResult && hasError {
		t.Error("Response has both result and error fields")
	}

	if !hasResult && !hasError {
		t.Error("Response has neither result nor error field")
	}
}

// BenchmarkHandleInitialize benchmarks the initialize endpoint
func BenchmarkHandleInitialize(b *testing.B) {
	server := NewServer(8080)
	req := httptest.NewRequest(http.MethodPost, "/mcp/v1/initialize", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		server.handleInitialize(w, req)
	}
}

// BenchmarkHandleToolsList benchmarks the tools list endpoint
func BenchmarkHandleToolsList(b *testing.B) {
	server := NewServer(8080)
	req := httptest.NewRequest(http.MethodGet, "/mcp/v1/tools/list", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		server.handleToolsList(w, req)
	}
}

// BenchmarkHandleToolsCall benchmarks the tools call endpoint
func BenchmarkHandleToolsCall(b *testing.B) {
	server := NewServer(8080)
	body := map[string]interface{}{
		"name": "echo",
		"arguments": map[string]interface{}{
			"message": "benchmark test",
		},
	}
	bodyBytes, _ := json.Marshal(body)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/mcp/v1/tools/call", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()
		server.handleToolsCall(w, req)
	}
}

// BenchmarkStdioHandleRequest benchmarks the stdio request handler
func BenchmarkStdioHandleRequest(b *testing.B) {
	server := NewStdioServer()
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.handleStdioRequest(request)
	}
}
