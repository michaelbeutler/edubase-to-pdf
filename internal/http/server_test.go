package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Helper function to add session_id as query parameter
func addSessionQuery(url, sessionID string) string {
	return url + "?session_id=" + sessionID
}

// TestNewServer tests server initialization
func TestNewServer(t *testing.T) {
	server := NewServer()
	assert.NotNil(t, server)
	assert.NotNil(t, server.sessions)
}

// TestCreateSession tests session creation
func TestCreateSession(t *testing.T) {
	server := NewServer()
	sessionID := server.CreateSession()
	assert.NotEmpty(t, sessionID)

	// Verify session exists
	session, exists := server.GetSession(sessionID)
	assert.True(t, exists)
	assert.NotNil(t, session)
	assert.Equal(t, sessionID, session.ID)
}

// TestGetSessionNotFound tests getting non-existent session
func TestGetSessionNotFound(t *testing.T) {
	server := NewServer()
	session, exists := server.GetSession("non-existent")
	assert.False(t, exists)
	assert.Nil(t, session)
}

// TestGetBooksHandlerNoSession tests GET /books without session
func TestGetBooksHandlerNoSession(t *testing.T) {
	server := NewServer()

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.GetBooksHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestGetBooksHandlerWithInvalidSession tests GET /books with invalid session
func TestGetBooksHandlerWithInvalidSession(t *testing.T) {
	server := NewServer()

	req := httptest.NewRequest(http.MethodGet, addSessionQuery("/books", "invalid-session-id"), nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.GetBooksHandler)
	handler.ServeHTTP(rr, req)

	// With auto-create, this should now succeed
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestLoginHandler tests the POST /login endpoint
func TestLoginHandlerNoSession(t *testing.T) {
	server := NewServer()

	requestBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.LoginHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestLoginHandlerInvalidRequest tests POST /login with invalid data
func TestLoginHandlerInvalidRequest(t *testing.T) {
	server := NewServer()
	sessionID := server.CreateSession()

	req := httptest.NewRequest(http.MethodPost, addSessionQuery("/api/login", sessionID), bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.LoginHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestLoginHandlerMissingCredentials tests POST /login with missing credentials
func TestLoginHandlerMissingCredentials(t *testing.T) {
	server := NewServer()
	sessionID := server.CreateSession()

	requestBody := LoginRequest{
		Email:    "",
		Password: "",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, addSessionQuery("/api/login", sessionID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.LoginHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestStartDownloadHandler tests the POST /download endpoint
func TestStartDownloadHandler(t *testing.T) {
	server := NewServer()
	sessionID := server.CreateSession()

	requestBody := StartDownloadRequest{
		BookID: 123,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, addSessionQuery("/download", sessionID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.StartDownloadHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)

	var response StartDownloadResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.JobID)
	assert.Equal(t, "pending", response.Status)
}

// TestStartDownloadHandlerNoSession tests POST /download without session
func TestStartDownloadHandlerNoSession(t *testing.T) {
	server := NewServer()

	requestBody := StartDownloadRequest{
		BookID: 123,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/download", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.StartDownloadHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestStartDownloadHandlerInvalidRequest tests POST /download with invalid data
func TestStartDownloadHandlerInvalidRequest(t *testing.T) {
	server := NewServer()
	sessionID := server.CreateSession()

	req := httptest.NewRequest(http.MethodPost, addSessionQuery("/download", sessionID), bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.StartDownloadHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestGetDownloadStatusHandler tests the GET /download/:jobId endpoint
func TestGetDownloadStatusHandler(t *testing.T) {
	server := NewServer()
	sessionID := server.CreateSession()

	// First start a download
	requestBody := StartDownloadRequest{
		BookID: 123,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, addSessionQuery("/download", sessionID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.StartDownloadHandler)
	handler.ServeHTTP(rr, req)

	var startResponse StartDownloadResponse
	json.NewDecoder(rr.Body).Decode(&startResponse)

	// Now get status
	req = httptest.NewRequest(http.MethodGet, addSessionQuery("/api/download/"+startResponse.JobID, sessionID), nil)

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(server.GetDownloadStatusHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response JobStatusResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, startResponse.JobID, response.JobID)
}

// TestGetDownloadStatusHandlerNotFound tests GET /download/:jobId for non-existent job
func TestGetDownloadStatusHandlerNotFound(t *testing.T) {
	server := NewServer()
	sessionID := server.CreateSession()

	req := httptest.NewRequest(http.MethodGet, addSessionQuery("/api/download/non-existent", sessionID), nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.GetDownloadStatusHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// TestDownloadPDFHandler tests the GET /download/:jobId/pdf endpoint
func TestDownloadPDFHandler(t *testing.T) {
	server := NewServer()
	sessionID := server.CreateSession()
	session, _ := server.GetSession(sessionID)

	// Create a completed job manually for testing
	jobID := "test-job-id"
	job := &DownloadJob{
		ID:          jobID,
		BookID:      123,
		Status:      "completed",
		Progress:    100,
		TotalPages:  10,
		Message:     "Download completed",
		PDFPath:     "/tmp/test.pdf",
		StartedAt:   time.Now(),
		CompletedAt: time.Now(),
	}
	session.jobs[jobID] = job

	req := httptest.NewRequest(http.MethodGet, addSessionQuery("/api/download/"+jobID+"/pdf", sessionID), nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.DownloadPDFHandler)
	handler.ServeHTTP(rr, req)

	// Since we don't have an actual file, we expect 404
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// TestDownloadPDFHandlerNotCompleted tests downloading PDF before job completion
func TestDownloadPDFHandlerNotCompleted(t *testing.T) {
	server := NewServer()
	sessionID := server.CreateSession()
	session, _ := server.GetSession(sessionID)

	// Create a pending job
	jobID := "test-job-id"
	job := &DownloadJob{
		ID:         jobID,
		BookID:     123,
		Status:     "pending",
		Progress:   0,
		TotalPages: 10,
		Message:    "Starting download",
		StartedAt:  time.Now(),
	}
	session.jobs[jobID] = job

	req := httptest.NewRequest(http.MethodGet, addSessionQuery("/api/download/"+jobID+"/pdf", sessionID), nil)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.DownloadPDFHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestSSEHandler tests the SSE /download/:jobId/events endpoint
func TestSSEHandler(t *testing.T) {
	server := NewServer()
	sessionID := server.CreateSession()

	// Start a download first
	requestBody := StartDownloadRequest{
		BookID: 123,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, addSessionQuery("/download", sessionID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.StartDownloadHandler)
	handler.ServeHTTP(rr, req)

	var startResponse StartDownloadResponse
	json.NewDecoder(rr.Body).Decode(&startResponse)

	// Now test SSE endpoint headers
	req = httptest.NewRequest(http.MethodGet, addSessionQuery("/api/download/"+startResponse.JobID+"/events", sessionID), nil)

	rr = httptest.NewRecorder()

	// We can't fully test SSE without a real connection, but we can test initial setup
	go func() {
		handler := http.HandlerFunc(server.SSEHandler)
		handler.ServeHTTP(rr, req)
	}()

	// Give it a moment to set headers
	time.Sleep(100 * time.Millisecond)

	// Verify SSE headers are set
	assert.Equal(t, "text/event-stream", rr.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", rr.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", rr.Header().Get("Connection"))
}

// TestSessionIsolation tests that sessions are properly isolated
func TestSessionIsolation(t *testing.T) {
	server := NewServer()

	// Create two sessions
	session1ID := server.CreateSession()
	session2ID := server.CreateSession()

	assert.NotEqual(t, session1ID, session2ID)

	// Add a job to session 1
	requestBody := StartDownloadRequest{
		BookID: 123,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, addSessionQuery("/download", session1ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.StartDownloadHandler)
	handler.ServeHTTP(rr, req)

	var startResponse StartDownloadResponse
	json.NewDecoder(rr.Body).Decode(&startResponse)

	// Try to access the job from session 2
	req = httptest.NewRequest(http.MethodGet, addSessionQuery("/download/"+startResponse.JobID, session2ID), nil)

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(server.GetDownloadStatusHandler)
	handler.ServeHTTP(rr, req)

	// Should not find the job in session 2
	assert.Equal(t, http.StatusNotFound, rr.Code)
}
