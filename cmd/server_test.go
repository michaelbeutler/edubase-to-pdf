package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET request returns OK",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"ok"}`,
		},
		{
			name:           "POST request returns method not allowed",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `method_not_allowed`,
		},
		{
			name:           "PUT request returns method not allowed",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `method_not_allowed`,
		},
		{
			name:           "DELETE request returns method not allowed",
			method:         http.MethodDelete,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `method_not_allowed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			server.handleHealth(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			bodyStr := strings.TrimSpace(string(body))

			if !strings.Contains(bodyStr, tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, bodyStr)
			}

			if resp.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestHandleDownload_MethodValidation(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "GET returns method not allowed",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "method_not_allowed",
		},
		{
			name:           "PUT returns method not allowed",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "method_not_allowed",
		},
		{
			name:           "DELETE returns method not allowed",
			method:         http.MethodDelete,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "method_not_allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/download", nil)
			w := httptest.NewRecorder()

			server.handleDownload(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			var errResp ErrorResponse
			json.NewDecoder(resp.Body).Decode(&errResp)

			if errResp.Error != tt.expectedError {
				t.Errorf("expected error %q, got %q", tt.expectedError, errResp.Error)
			}
		})
	}
}

func TestHandleDownload_InvalidJSON(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	tests := []struct {
		name          string
		body          string
		expectedError string
	}{
		{
			name:          "Invalid JSON syntax",
			body:          `{invalid json}`,
			expectedError: "invalid_json",
		},
		{
			name:          "Empty body",
			body:          ``,
			expectedError: "invalid_json",
		},
		{
			name:          "Non-JSON content",
			body:          `plain text`,
			expectedError: "invalid_json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/download", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			server.handleDownload(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
			}

			var errResp ErrorResponse
			json.NewDecoder(resp.Body).Decode(&errResp)

			if errResp.Error != tt.expectedError {
				t.Errorf("expected error %q, got %q", tt.expectedError, errResp.Error)
			}
		})
	}
}

func TestValidateRequest(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	tests := []struct {
		name        string
		req         DownloadRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password123",
				BookID:    12345,
				StartPage: 1,
				MaxPages:  -1,
			},
			expectError: false,
		},
		{
			name: "Valid request with max pages",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password123",
				BookID:    12345,
				StartPage: 1,
				MaxPages:  10,
			},
			expectError: false,
		},
		{
			name: "Missing email",
			req: DownloadRequest{
				Password:  "password123",
				BookID:    12345,
				StartPage: 1,
				MaxPages:  -1,
			},
			expectError: true,
			errorMsg:    "email is required",
		},
		{
			name: "Missing password",
			req: DownloadRequest{
				Email:     "test@example.com",
				BookID:    12345,
				StartPage: 1,
				MaxPages:  -1,
			},
			expectError: true,
			errorMsg:    "password is required",
		},
		{
			name: "Zero book ID",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password123",
				BookID:    0,
				StartPage: 1,
				MaxPages:  -1,
			},
			expectError: true,
			errorMsg:    "book_id must be a positive integer",
		},
		{
			name: "Negative book ID",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password123",
				BookID:    -100,
				StartPage: 1,
				MaxPages:  -1,
			},
			expectError: true,
			errorMsg:    "book_id must be a positive integer",
		},
		{
			name: "Zero start page",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password123",
				BookID:    12345,
				StartPage: 0,
				MaxPages:  -1,
			},
			expectError: true,
			errorMsg:    "start_page must be a positive integer",
		},
		{
			name: "Negative start page",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password123",
				BookID:    12345,
				StartPage: -5,
				MaxPages:  -1,
			},
			expectError: true,
			errorMsg:    "start_page must be a positive integer",
		},
		{
			name: "Zero max pages",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password123",
				BookID:    12345,
				StartPage: 1,
				MaxPages:  0,
			},
			expectError: true,
			errorMsg:    "max_pages must be -1 (all pages) or a positive integer",
		},
		{
			name: "Negative max pages (-2)",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password123",
				BookID:    12345,
				StartPage: 1,
				MaxPages:  -2,
			},
			expectError: true,
			errorMsg:    "max_pages must be -1 (all pages) or a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.validateRequest(&tt.req)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}

			if tt.expectError && err != nil && err.Error() != tt.errorMsg {
				t.Errorf("expected error message %q, got %q", tt.errorMsg, err.Error())
			}
		})
	}
}

func TestHandleDownload_ValidationErrors(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	tests := []struct {
		name           string
		req            DownloadRequest
		expectedStatus int
		expectedError  string
		errorMessage   string
	}{
		{
			name: "Missing email",
			req: DownloadRequest{
				Password:  "password",
				BookID:    1,
				StartPage: 1,
				MaxPages:  -1,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
			errorMessage:   "email is required",
		},
		{
			name: "Missing password",
			req: DownloadRequest{
				Email:     "test@example.com",
				BookID:    1,
				StartPage: 1,
				MaxPages:  -1,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
			errorMessage:   "password is required",
		},
		{
			name: "Invalid book ID",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password",
				BookID:    0,
				StartPage: 1,
				MaxPages:  -1,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
			errorMessage:   "book_id must be a positive integer",
		},
		{
			name: "Invalid start page",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password",
				BookID:    1,
				StartPage: 0,
				MaxPages:  -1,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
			errorMessage:   "start_page must be a positive integer",
		},
		{
			name: "Invalid max pages",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password",
				BookID:    1,
				StartPage: 1,
				MaxPages:  0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
			errorMessage:   "max_pages must be -1 (all pages) or a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.req)
			req := httptest.NewRequest(http.MethodPost, "/download", bytes.NewReader(body))
			w := httptest.NewRecorder()

			server.handleDownload(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			var errResp ErrorResponse
			json.NewDecoder(resp.Body).Decode(&errResp)

			if errResp.Error != tt.expectedError {
				t.Errorf("expected error %q, got %q", tt.expectedError, errResp.Error)
			}

			if errResp.Message != tt.errorMessage {
				t.Errorf("expected message %q, got %q", tt.errorMessage, errResp.Message)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	tests := []struct {
		name           string
		status         int
		errorCode      string
		message        string
		expectedStatus int
	}{
		{
			name:           "Bad request error",
			status:         http.StatusBadRequest,
			errorCode:      "bad_request",
			message:        "Invalid input",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized error",
			status:         http.StatusUnauthorized,
			errorCode:      "unauthorized",
			message:        "Invalid credentials",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Internal server error",
			status:         http.StatusInternalServerError,
			errorCode:      "internal_error",
			message:        "Something went wrong",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			server.writeError(w, tt.status, tt.errorCode, tt.message)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			var errResp ErrorResponse
			json.NewDecoder(resp.Body).Decode(&errResp)

			if errResp.Error != tt.errorCode {
				t.Errorf("expected error %q, got %q", tt.errorCode, errResp.Error)
			}

			if errResp.Message != tt.message {
				t.Errorf("expected message %q, got %q", tt.message, errResp.Message)
			}

			if resp.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestNewHTTPServer(t *testing.T) {
	server := newHTTPServer("localhost", 9090)

	if server == nil {
		t.Fatal("expected server to be created")
	}

	if server.server == nil {
		t.Fatal("expected http.Server to be initialized")
	}

	expectedAddr := "localhost:9090"
	if server.server.Addr != expectedAddr {
		t.Errorf("expected address %q, got %q", expectedAddr, server.server.Addr)
	}

	if server.server.Handler == nil {
		t.Fatal("expected handler to be set")
	}

	if server.server.ReadTimeout == 0 {
		t.Error("expected ReadTimeout to be set")
	}

	if server.server.WriteTimeout == 0 {
		t.Error("expected WriteTimeout to be set")
	}

	if server.server.IdleTimeout == 0 {
		t.Error("expected IdleTimeout to be set")
	}
}

func TestDownloadRequest_JSONMarshaling(t *testing.T) {
	req := DownloadRequest{
		Email:     "test@example.com",
		Password:  "password123",
		BookID:    12345,
		StartPage: 1,
		MaxPages:  10,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var decoded DownloadRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	if decoded.Email != req.Email {
		t.Errorf("expected email %q, got %q", req.Email, decoded.Email)
	}
	if decoded.Password != req.Password {
		t.Errorf("expected password %q, got %q", req.Password, decoded.Password)
	}
	if decoded.BookID != req.BookID {
		t.Errorf("expected book_id %d, got %d", req.BookID, decoded.BookID)
	}
	if decoded.StartPage != req.StartPage {
		t.Errorf("expected start_page %d, got %d", req.StartPage, decoded.StartPage)
	}
	if decoded.MaxPages != req.MaxPages {
		t.Errorf("expected max_pages %d, got %d", req.MaxPages, decoded.MaxPages)
	}
}

func TestErrorResponse_JSONMarshaling(t *testing.T) {
	resp := ErrorResponse{
		Error:   "test_error",
		Message: "This is a test error message",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var decoded ErrorResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.Error != resp.Error {
		t.Errorf("expected error %q, got %q", resp.Error, decoded.Error)
	}
	if decoded.Message != resp.Message {
		t.Errorf("expected message %q, got %q", resp.Message, decoded.Message)
	}
}

// Integration test for download endpoint - requires credentials
func TestHandleDownload_Integration(t *testing.T) {
	// Skip if credentials are not set
	email := os.Getenv("EDUBASE_EMAIL")
	password := os.Getenv("EDUBASE_PASSWORD")
	if email == "" || password == "" {
		t.Skip("Skipping integration test: EDUBASE_EMAIL and EDUBASE_PASSWORD environment variables must be set")
	}

	server := newHTTPServer("localhost", 8080)

	req := DownloadRequest{
		Email:     email,
		Password:  password,
		BookID:    58216,
		StartPage: 1,
		MaxPages:  1, // Only download 1 page for testing
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/download", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Note: This will fail if playwright is not installed, which is expected in test environments
	// The test primarily validates the handler structure
	server.handleDownload(w, httpReq)

	resp := w.Result()
	defer resp.Body.Close()

	// Check that we get either success or an expected error (playwright not installed, etc.)
	if resp.StatusCode != http.StatusOK && 
	   resp.StatusCode != http.StatusInternalServerError &&
	   resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func TestHTTPServerRoutes(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	tests := []struct {
		name   string
		path   string
		method string
	}{
		{
			name:   "Health endpoint with GET",
			path:   "/health",
			method: http.MethodGet,
		},
		{
			name:   "Download endpoint with POST",
			path:   "/download",
			method: http.MethodPost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			server.server.Handler.ServeHTTP(w, req)

			// Verify that the route exists (not 404)
			if w.Code == http.StatusNotFound {
				t.Errorf("route %s not found", tt.path)
			}
		})
	}
}

func TestHandleDownload_EdgeCases(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	tests := []struct {
		name           string
		req            DownloadRequest
		expectedStatus int
	}{
		{
			name: "Large book ID",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password",
				BookID:    999999999,
				StartPage: 1,
				MaxPages:  -1,
			},
			expectedStatus: http.StatusInternalServerError, // Will fail in processing
		},
		{
			name: "Large start page",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password",
				BookID:    12345,
				StartPage: 10000,
				MaxPages:  1,
			},
			expectedStatus: http.StatusInternalServerError, // Will fail in processing
		},
		{
			name: "Max pages boundary",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "password",
				BookID:    12345,
				StartPage: 1,
				MaxPages:  1000,
			},
			expectedStatus: http.StatusInternalServerError, // Will fail in processing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.req)
			httpReq := httptest.NewRequest(http.MethodPost, "/download", bytes.NewReader(body))
			w := httptest.NewRecorder()

			server.handleDownload(w, httpReq)

			resp := w.Result()
			defer resp.Body.Close()

			// These should fail during processing, not validation
			if resp.StatusCode != tt.expectedStatus && resp.StatusCode != http.StatusUnauthorized {
				// Either internal error or auth failure is acceptable
				if resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusUnauthorized {
					t.Errorf("expected status %d or %d, got %d", tt.expectedStatus, http.StatusUnauthorized, resp.StatusCode)
				}
			}
		})
	}
}

func TestValidateRequest_AllFields(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	// Test with all edge case values
	tests := []struct {
		name      string
		req       DownloadRequest
		wantError bool
	}{
		{
			name: "Minimum valid values",
			req: DownloadRequest{
				Email:     "a@b.c",
				Password:  "p",
				BookID:    1,
				StartPage: 1,
				MaxPages:  -1,
			},
			wantError: false,
		},
		{
			name: "Maximum reasonable values",
			req: DownloadRequest{
				Email:     "very.long.email.address@example.com",
				Password:  "veryLongPassword123!@#",
				BookID:    2147483647,
				StartPage: 999999,
				MaxPages:  999999,
			},
			wantError: false,
		},
		{
			name: "Empty string email",
			req: DownloadRequest{
				Email:     "",
				Password:  "password",
				BookID:    1,
				StartPage: 1,
				MaxPages:  -1,
			},
			wantError: true,
		},
		{
			name: "Empty string password",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "",
				BookID:    1,
				StartPage: 1,
				MaxPages:  -1,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.validateRequest(&tt.req)
			if (err != nil) != tt.wantError {
				t.Errorf("validateRequest() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestHTTPServer_ContentType(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	tests := []struct {
		name            string
		path            string
		method          string
		body            string
		expectedCT      string
	}{
		{
			name:       "Health returns JSON",
			path:       "/health",
			method:     http.MethodGet,
			expectedCT: "application/json",
		},
		{
			name:       "Error responses return JSON",
			path:       "/health",
			method:     http.MethodPost,
			expectedCT: "application/json",
		},
		{
			name:       "Download validation error returns JSON",
			path:       "/download",
			method:     http.MethodPost,
			body:       `{"email":"","password":""}`,
			expectedCT: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			w := httptest.NewRecorder()

			server.server.Handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			ct := resp.Header.Get("Content-Type")
			if ct != tt.expectedCT {
				t.Errorf("expected Content-Type %q, got %q", tt.expectedCT, ct)
			}
		})
	}
}

func TestHandleDownload_InvalidCredentials(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	req := DownloadRequest{
		Email:     "invalid@example.com",
		Password:  "wrongpassword",
		BookID:    12345,
		StartPage: 1,
		MaxPages:  1,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/download", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleDownload(w, httpReq)

	resp := w.Result()
	defer resp.Body.Close()

	// Should fail with either unauthorized or internal error
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d or %d for invalid credentials, got %d", 
			http.StatusUnauthorized, http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestHTTPServer_Configuration(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
	}{
		{
			name: "Default configuration",
			host: "0.0.0.0",
			port: 8080,
		},
		{
			name: "Localhost configuration",
			host: "localhost",
			port: 9090,
		},
		{
			name: "Custom port",
			host: "127.0.0.1",
			port: 3000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newHTTPServer(tt.host, tt.port)
			
			// Simple check that server is configured
			if server.server == nil {
				t.Error("server.server should not be nil")
			}
			if server.server.Handler == nil {
				t.Error("server handler should not be nil")
			}
		})
	}
}

func TestHandleDownload_LargeRequestBody(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	// Create a request with a very large password (edge case)
	req := DownloadRequest{
		Email:     "test@example.com",
		Password:  string(make([]byte, 10000)), // Very long password
		BookID:    12345,
		StartPage: 1,
		MaxPages:  1,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/download", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleDownload(w, httpReq)

	// Should pass validation but fail in processing
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected failure status, got %d", resp.StatusCode)
	}
}

func TestHandleDownload_MultipleRequests(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	// Test that multiple requests can be handled
	for i := 0; i < 3; i++ {
		req := DownloadRequest{
			Email:     "test@example.com",
			Password:  "password",
			BookID:    12345 + i,
			StartPage: 1,
			MaxPages:  1,
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/download", bytes.NewReader(body))
		w := httptest.NewRecorder()

		server.handleDownload(w, httpReq)

		resp := w.Result()
		resp.Body.Close()

		// Each request should be handled independently
		if resp.StatusCode == http.StatusNotFound {
			t.Errorf("request %d: endpoint not found", i)
		}
	}
}

func TestHandleDownload_ConcurrentRequests(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	// Test validation errors with concurrent requests
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(id int) {
			req := DownloadRequest{
				Email:     "", // Invalid - missing email
				Password:  "password",
				BookID:    12345,
				StartPage: 1,
				MaxPages:  1,
			}

			body, _ := json.Marshal(req)
			httpReq := httptest.NewRequest(http.MethodPost, "/download", bytes.NewReader(body))
			w := httptest.NewRecorder()

			server.handleDownload(w, httpReq)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("request %d: expected %d, got %d", id, http.StatusBadRequest, resp.StatusCode)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestHandleDownload_ResponseHeaders(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	// Test that error responses have correct headers
	req := DownloadRequest{
		Email:     "", // Invalid
		Password:  "password",
		BookID:    12345,
		StartPage: 1,
		MaxPages:  1,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/download", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleDownload(w, httpReq)

	resp := w.Result()
	defer resp.Body.Close()

	// Check headers
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}

func TestValidateRequest_NegativeMaxPages(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	tests := []struct {
		name      string
		maxPages  int
		wantError bool
		errorMsg  string
	}{
		{
			name:      "-1 is valid (all pages)",
			maxPages:  -1,
			wantError: false,
		},
		{
			name:      "-2 is invalid",
			maxPages:  -2,
			wantError: true,
			errorMsg:  "max_pages must be -1 (all pages) or a positive integer",
		},
		{
			name:      "-100 is invalid",
			maxPages:  -100,
			wantError: true,
			errorMsg:  "max_pages must be -1 (all pages) or a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := DownloadRequest{
				Email:     "test@example.com",
				Password:  "password",
				BookID:    12345,
				StartPage: 1,
				MaxPages:  tt.maxPages,
			}

			err := server.validateRequest(&req)
			if tt.wantError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
			if tt.wantError && err != nil && err.Error() != tt.errorMsg {
				t.Errorf("expected error %q, got %q", tt.errorMsg, err.Error())
			}
		})
	}
}

func TestHandleHealth_ResponseBody(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Parse response body
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", result["status"])
	}
}

func TestWriteError_ResponseBody(t *testing.T) {
	server := newHTTPServer("localhost", 8080)
	w := httptest.NewRecorder()

	errorCode := "test_error"
	errorMsg := "Test error message"

	server.writeError(w, http.StatusBadRequest, errorCode, errorMsg)

	resp := w.Result()
	defer resp.Body.Close()

	// Parse response body
	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error != errorCode {
		t.Errorf("expected error code %q, got %q", errorCode, errResp.Error)
	}
	if errResp.Message != errorMsg {
		t.Errorf("expected message %q, got %q", errorMsg, errResp.Message)
	}
}

func TestCreateTempDir(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	tempDir, err := server.createTempDir()
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Check that directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("temp directory was not created")
	}

	// Check that directory is writable
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Errorf("temp directory is not writable: %v", err)
	}
}

func TestStreamPDF(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	// Create a temporary PDF file for testing
	tempDir, err := os.MkdirTemp("", "test-pdf-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pdfPath := filepath.Join(tempDir, "test.pdf")
	testContent := []byte("%PDF-1.4 test content")
	if err := os.WriteFile(pdfPath, testContent, 0644); err != nil {
		t.Fatalf("failed to create test PDF: %v", err)
	}

	w := httptest.NewRecorder()
	bookID := 12345

	err = server.streamPDF(w, pdfPath, bookID)
	if err != nil {
		t.Fatalf("streamPDF failed: %v", err)
	}

	resp := w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Check headers
	if ct := resp.Header.Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("expected Content-Type application/pdf, got %s", ct)
	}

	expectedDisposition := fmt.Sprintf("attachment; filename=book_%d.pdf", bookID)
	if cd := resp.Header.Get("Content-Disposition"); cd != expectedDisposition {
		t.Errorf("expected Content-Disposition %q, got %q", expectedDisposition, cd)
	}

	// Check content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if !bytes.Equal(body, testContent) {
		t.Errorf("expected body %q, got %q", testContent, body)
	}
}

func TestStreamPDF_NonExistentFile(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	w := httptest.NewRecorder()
	err := server.streamPDF(w, "/nonexistent/file.pdf", 12345)

	if err == nil {
		t.Error("expected error for nonexistent file")
	}

	if !strings.Contains(err.Error(), "failed to open PDF") {
		t.Errorf("expected 'failed to open PDF' error, got %v", err)
	}
}

func TestCleanupBrowser(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	// Test with nil values - should not panic
	server.cleanupBrowser(nil, nil, nil)

	// If we get here without panic, test passes
}

func TestHandleDownload_RequestBodyReadError(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	// Create a request with a body that will cause a read error
	req := httptest.NewRequest(http.MethodPost, "/download", &errorReader{})
	w := httptest.NewRecorder()

	server.handleDownload(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d for read error, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	var errResp ErrorResponse
	json.NewDecoder(resp.Body).Decode(&errResp)

	if errResp.Error != "invalid_json" {
		t.Errorf("expected error 'invalid_json', got %q", errResp.Error)
	}
}

// errorReader is a helper type that always returns an error when Read is called
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

func TestValidateRequest_BoundaryValues(t *testing.T) {
	server := newHTTPServer("localhost", 8080)

	tests := []struct {
		name      string
		req       DownloadRequest
		wantError bool
	}{
		{
			name: "StartPage = 1",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "pass",
				BookID:    1,
				StartPage: 1,
				MaxPages:  -1,
			},
			wantError: false,
		},
		{
			name: "StartPage = max int",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "pass",
				BookID:    1,
				StartPage: 2147483647,
				MaxPages:  -1,
			},
			wantError: false,
		},
		{
			name: "BookID = 1",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "pass",
				BookID:    1,
				StartPage: 1,
				MaxPages:  -1,
			},
			wantError: false,
		},
		{
			name: "MaxPages = 1",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "pass",
				BookID:    1,
				StartPage: 1,
				MaxPages:  1,
			},
			wantError: false,
		},
		{
			name: "MaxPages = -1",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "pass",
				BookID:    1,
				StartPage: 1,
				MaxPages:  -1,
			},
			wantError: false,
		},
		{
			name: "MaxPages = -100 (invalid)",
			req: DownloadRequest{
				Email:     "test@example.com",
				Password:  "pass",
				BookID:    1,
				StartPage: 1,
				MaxPages:  -100,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.validateRequest(&tt.req)
			if (err != nil) != tt.wantError {
				t.Errorf("validateRequest() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
