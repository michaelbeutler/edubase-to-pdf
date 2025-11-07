package http

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/michaelbeutler/edubase-to-pdf/pkg/edubase"
	"github.com/playwright-community/playwright-go"
	"golang.org/x/image/draw"
)

// Server represents the HTTP server with session management
type Server struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// Session represents a user session with authentication and state
type Session struct {
	ID          string
	jobs        map[string]*DownloadJob
	jobsMu      sync.RWMutex
	page        playwright.Page
	browser     playwright.Browser
	playwright  *playwright.Playwright
	credentials edubase.Credentials
	createdAt   time.Time
}

// DownloadJob represents an ongoing or completed book download
type DownloadJob struct {
	ID           string
	BookID       int
	Width        int
	Height       int
	Status       string // pending, downloading, completed, failed
	Progress     int    // current page number
	TotalPages   int
	Message      string
	PDFPath      string
	StartedAt    time.Time
	CompletedAt  time.Time
	Error        string
	eventClients []chan ProgressEvent
	clientsMu    sync.RWMutex
}

// ProgressEvent represents a real-time progress update
type ProgressEvent struct {
	JobID      string    `json:"job_id"`
	Status     string    `json:"status"`
	Progress   int       `json:"progress"`
	TotalPages int       `json:"total_pages"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
}

// Request/Response types
type BooksResponse struct {
	Books []edubase.Book `json:"books"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type StartDownloadRequest struct {
	BookID int `json:"book_id"`
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

type StartDownloadResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

type JobStatusResponse struct {
	JobID      string    `json:"job_id"`
	BookID     int       `json:"book_id"`
	Status     string    `json:"status"`
	Progress   int       `json:"progress"`
	TotalPages int       `json:"total_pages"`
	Message    string    `json:"message"`
	Error      string    `json:"error,omitempty"`
	StartedAt  time.Time `json:"started_at"`
}

// NewServer creates a new HTTP server instance
func NewServer() *Server {
	return &Server{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session and returns the session ID
func (s *Server) CreateSession() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID := uuid.New().String()
	session := &Session{
		ID:        sessionID,
		jobs:      make(map[string]*DownloadJob),
		createdAt: time.Now(),
	}
	s.sessions[sessionID] = session

	return sessionID
}

// GetSession retrieves a session by ID
func (s *Server) GetSession(sessionID string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	return session, exists
}

// getSessionFromRequest extracts session ID from query param and retrieves or creates the session
func (s *Server) getSessionFromRequest(r *http.Request) (*Session, error) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		return nil, fmt.Errorf("session_id query parameter required")
	}

	// Try to get existing session
	session, exists := s.GetSession(sessionID)
	if exists {
		return session, nil
	}

	// Auto-create session if it doesn't exist
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring lock
	session, exists = s.sessions[sessionID]
	if exists {
		return session, nil
	}

	// Create new session with the provided ID
	session = &Session{
		ID:        sessionID,
		jobs:      make(map[string]*DownloadJob),
		createdAt: time.Now(),
	}
	s.sessions[sessionID] = session

	return session, nil
}

// GetBooksHandler handles GET /books requests
func (s *Server) GetBooksHandler(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Check if page is initialized
	if session.page == nil {
		http.Error(w, "Session not authenticated. Please login first.", http.StatusUnauthorized)
		return
	}

	// Create library provider
	libraryProvider := edubase.NewLibraryProvider(session.page)
	books, err := libraryProvider.GetBooks()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get books: %v", err), http.StatusInternalServerError)
		return
	}

	response := BooksResponse{Books: books}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LoginHandler handles POST /login requests
func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate credentials
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Initialize Playwright if not already done
	if session.playwright == nil {
		pw, err := playwright.Run()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to start playwright: %v", err), http.StatusInternalServerError)
			return
		}
		session.playwright = pw
	}

	// Launch browser if not already done
	if session.browser == nil {
		browser, err := session.playwright.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(true),
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to launch browser: %v", err), http.StatusInternalServerError)
			return
		}
		session.browser = browser
	}

	// Create new page if not already done
	if session.page == nil {
		page, err := session.browser.NewPage()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create page: %v", err), http.StatusInternalServerError)
			return
		}
		session.page = page
	}

	// Store credentials
	session.credentials = edubase.Credentials{
		Email:    req.Email,
		Password: req.Password,
	}

	// Perform login
	loginProvider := edubase.NewLoginProvider(session.page)
	if err := loginProvider.Login(session.credentials, false); err != nil {
		response := LoginResponse{
			Success: false,
			Message: fmt.Sprintf("Login failed: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := LoginResponse{
		Success: true,
		Message: "Login successful",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartDownloadHandler handles POST /download requests
func (s *Server) StartDownloadHandler(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req StartDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create new job
	jobID := uuid.New().String()

	// Set default resolution if not provided (4K)
	width := req.Width
	height := req.Height
	if width <= 0 {
		width = 3840 // 4K default
	}
	if height <= 0 {
		height = 2160 // 4K default
	}

	job := &DownloadJob{
		ID:           jobID,
		BookID:       req.BookID,
		Width:        width,
		Height:       height,
		Status:       "pending",
		Progress:     0,
		Message:      fmt.Sprintf("Download queued (Resolution: %dx%d)", width, height),
		StartedAt:    time.Now(),
		eventClients: make([]chan ProgressEvent, 0),
	}

	session.jobsMu.Lock()
	session.jobs[jobID] = job
	session.jobsMu.Unlock()

	// Start download in background
	go s.processDownload(session, job)

	response := StartDownloadResponse{
		JobID:  jobID,
		Status: "pending",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

// processDownload handles the actual book download process
func (s *Server) processDownload(session *Session, job *DownloadJob) {
	// Skip actual download if page is nil (testing mode)
	if session.page == nil {
		job.Status = "failed"
		job.Error = "Session not initialized with playwright page"
		job.Message = "Session not properly initialized"
		s.broadcastProgress(job)
		return
	}

	job.Status = "downloading"
	job.Message = "Starting download"
	s.broadcastProgress(job)

	// Set viewport size for better quality screenshots
	if err := session.page.SetViewportSize(job.Width, job.Height); err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("Failed to set viewport size: %v", err)
		job.Message = "Failed to set viewport size"
		s.broadcastProgress(job)
		return
	}

	// Create book provider
	bookProvider := edubase.NewBookProvider(session.page, job.BookID)

	// Open book
	if err := bookProvider.Open(1); err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("Failed to open book: %v", err)
		job.Message = "Failed to open book"
		s.broadcastProgress(job)
		return
	}

	// Get total pages
	totalPages, err := bookProvider.GetTotalPages()
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("Failed to get total pages: %v", err)
		job.Message = "Failed to get total pages"
		s.broadcastProgress(job)
		return
	}

	job.TotalPages = totalPages
	job.Message = fmt.Sprintf("Downloading %d pages", totalPages)
	s.broadcastProgress(job)

	// Create temp directory for screenshots
	tempDir := filepath.Join(os.TempDir(), "edubase-downloads", job.ID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("Failed to create temp directory: %v", err)
		job.Message = "Failed to create temp directory"
		s.broadcastProgress(job)
		return
	}

	// Download each page
	ocrTexts := make([]string, totalPages)

	for i := 1; i <= totalPages; i++ {
		job.Progress = i
		job.Message = fmt.Sprintf("Downloading page %d of %d", i, totalPages)
		s.broadcastProgress(job)

		// Extract text from the page BEFORE taking screenshot
		pageText, err := bookProvider.GetPageText()
		if err != nil {
			fmt.Printf("Warning: Failed to extract text from page %d: %v\n", i, err)
			ocrTexts[i-1] = ""
		} else {
			fmt.Printf("Extracted %d characters from page %d\n", len(pageText), i)
			ocrTexts[i-1] = pageText
		}

		filename := filepath.Join(tempDir, fmt.Sprintf("page_%03d.jpg", i))
		if err := bookProvider.Screenshot(filename); err != nil {
			job.Status = "failed"
			job.Error = fmt.Sprintf("Failed to screenshot page %d: %v", i, err)
			job.Message = fmt.Sprintf("Failed to download page %d", i)
			s.broadcastProgress(job)
			return
		}

		// Go to next page if not the last page
		if i < totalPages {
			if err := bookProvider.NextPage(); err != nil {
				job.Status = "failed"
				job.Error = fmt.Sprintf("Failed to go to next page: %v", err)
				job.Message = "Failed to navigate to next page"
				s.broadcastProgress(job)
				return
			}
		}

		// Small delay between pages
		time.Sleep(500 * time.Millisecond)
	}

	// Create PDF from downloaded images
	job.Message = "Creating PDF from images"
	s.broadcastProgress(job)

	pdfPath := filepath.Join(tempDir, fmt.Sprintf("book_%d.pdf", job.BookID))
	a4Dir := filepath.Join(tempDir, "a4")

	// Create A4 directory for converted images
	if err := os.MkdirAll(a4Dir, 0755); err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("Failed to create A4 directory: %v", err)
		job.Message = "Failed to create A4 directory"
		s.broadcastProgress(job)
		return
	}

	// Convert images to A4 format
	job.Message = "Converting images to A4 format"
	s.broadcastProgress(job)

	var a4ImagePaths []string

	for i := 1; i <= totalPages; i++ {
		job.Progress = i
		job.Message = fmt.Sprintf("Converting to A4: page %d of %d", i, totalPages)
		s.broadcastProgress(job)

		srcPath := filepath.Join(tempDir, fmt.Sprintf("page_%03d.jpg", i))
		dstPath := filepath.Join(a4Dir, fmt.Sprintf("a4_page_%03d.jpg", i))

		if err := convertToA4(srcPath, dstPath); err != nil {
			job.Status = "failed"
			job.Error = fmt.Sprintf("Failed to convert page %d to A4: %v", i, err)
			job.Message = fmt.Sprintf("Failed to convert page %d", i)
			s.broadcastProgress(job)
			return
		}

		a4ImagePaths = append(a4ImagePaths, dstPath)
	}

	// Create PDF with A4 page size and searchable text
	job.Message = "Generating PDF with A4 pages and searchable text"
	s.broadcastProgress(job)

	if err := createSearchablePDF(pdfPath, a4ImagePaths, ocrTexts); err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("Failed to create PDF: %v", err)
		job.Message = "Failed to create PDF"
		s.broadcastProgress(job)
		return
	}

	job.PDFPath = pdfPath
	job.Status = "completed"
	job.Message = fmt.Sprintf("Download completed - %d pages in A4 format with searchable text", totalPages)
	job.CompletedAt = time.Now()
	s.broadcastProgress(job)
}

// convertToA4 converts an image to A4 format (595x842 points = 2480x3508 pixels at 300 DPI)
func convertToA4(srcPath, dstPath string) error {
	// Open source image
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source image: %v", err)
	}
	defer srcFile.Close()

	srcImg, _, err := image.Decode(srcFile)
	if err != nil {
		return fmt.Errorf("failed to decode source image: %v", err)
	}

	// A4 dimensions at 300 DPI (good print quality)
	const a4Width = 2480
	const a4Height = 3508

	// Calculate scaling to fit within A4 while maintaining aspect ratio
	srcBounds := srcImg.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	scaleX := float64(a4Width) / float64(srcWidth)
	scaleY := float64(a4Height) / float64(srcHeight)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	newWidth := int(float64(srcWidth) * scale)
	newHeight := int(float64(srcHeight) * scale)

	// Create A4 canvas (white background)
	a4Img := image.NewRGBA(image.Rect(0, 0, a4Width, a4Height))

	// Fill with white background
	for y := 0; y < a4Height; y++ {
		for x := 0; x < a4Width; x++ {
			a4Img.Set(x, y, image.White)
		}
	}

	// Calculate position to center the scaled image
	offsetX := (a4Width - newWidth) / 2
	offsetY := (a4Height - newHeight) / 2

	// Scale and draw the source image centered on A4 canvas
	dstRect := image.Rect(offsetX, offsetY, offsetX+newWidth, offsetY+newHeight)
	draw.CatmullRom.Scale(a4Img, dstRect, srcImg, srcBounds, draw.Over, nil)

	// Save the result
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	if err := jpeg.Encode(dstFile, a4Img, &jpeg.Options{Quality: 95}); err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	return nil
}

// broadcastProgress sends progress updates to all SSE clients
func (s *Server) broadcastProgress(job *DownloadJob) {
	event := ProgressEvent{
		JobID:      job.ID,
		Status:     job.Status,
		Progress:   job.Progress,
		TotalPages: job.TotalPages,
		Message:    job.Message,
		Timestamp:  time.Now(),
	}

	job.clientsMu.RLock()
	defer job.clientsMu.RUnlock()

	for _, client := range job.eventClients {
		select {
		case client <- event:
		default:
			// Client not ready, skip
		}
	}
}

// GetDownloadStatusHandler handles GET /download/:jobId requests
func (s *Server) GetDownloadStatusHandler(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Extract jobId from URL path (already trimmed by router)
	jobID := strings.TrimPrefix(r.URL.Path, "/api/download/")

	session.jobsMu.RLock()
	job, exists := session.jobs[jobID]
	session.jobsMu.RUnlock()

	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	response := JobStatusResponse{
		JobID:      job.ID,
		BookID:     job.BookID,
		Status:     job.Status,
		Progress:   job.Progress,
		TotalPages: job.TotalPages,
		Message:    job.Message,
		Error:      job.Error,
		StartedAt:  job.StartedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DownloadPDFHandler handles GET /download/:jobId/pdf requests
func (s *Server) DownloadPDFHandler(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Extract jobId from URL path (already trimmed by router)
	path := strings.TrimPrefix(r.URL.Path, "/api/download/")
	jobID := strings.TrimSuffix(path, "/pdf")

	session.jobsMu.RLock()
	job, exists := session.jobs[jobID]
	session.jobsMu.RUnlock()

	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	if job.Status != "completed" {
		http.Error(w, "Job not completed yet", http.StatusBadRequest)
		return
	}

	// Serve the PDF file
	if _, err := os.Stat(job.PDFPath); os.IsNotExist(err) {
		http.Error(w, "PDF file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=book_%d.pdf", job.BookID))
	http.ServeFile(w, r, job.PDFPath)
}

// SSEHandler handles GET /download/:jobId/events SSE requests
func (s *Server) SSEHandler(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSessionFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Extract jobId from URL path (already trimmed by router)
	path := strings.TrimPrefix(r.URL.Path, "/api/download/")
	jobID := strings.TrimSuffix(path, "/events")

	session.jobsMu.RLock()
	job, exists := session.jobs[jobID]
	session.jobsMu.RUnlock()

	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create event channel for this client
	eventChan := make(chan ProgressEvent, 10)

	job.clientsMu.Lock()
	job.eventClients = append(job.eventClients, eventChan)
	job.clientsMu.Unlock()

	// Remove client when done
	defer func() {
		job.clientsMu.Lock()
		for i, ch := range job.eventClients {
			if ch == eventChan {
				job.eventClients = append(job.eventClients[:i], job.eventClients[i+1:]...)
				break
			}
		}
		job.clientsMu.Unlock()
		close(eventChan)
	}()

	// Send initial status
	initialEvent := ProgressEvent{
		JobID:      job.ID,
		Status:     job.Status,
		Progress:   job.Progress,
		TotalPages: job.TotalPages,
		Message:    job.Message,
		Timestamp:  time.Now(),
	}

	data, _ := json.Marshal(initialEvent)
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Stream events
	for {
		select {
		case event := <-eventChan:
			data, err := json.Marshal(event)
			if err != nil {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			// Close connection if job is completed or failed
			if event.Status == "completed" || event.Status == "failed" {
				return
			}

		case <-r.Context().Done():
			return
		}
	}
}

// createSearchablePDF creates a PDF with images and invisible searchable text layer using gofpdf
func createSearchablePDF(pdfPath string, imagePaths []string, texts []string) error {
	fmt.Printf("Creating searchable PDF with %d images and %d text extracts\n", len(imagePaths), len(texts))

	pdf := gofpdf.New("P", "mm", "A4", "")

	// A4 dimensions in mm
	const pageWidth = 210.0
	const pageHeight = 297.0

	for i, imgPath := range imagePaths {
		// Add a new page
		pdf.AddPage()

		// Add the image - it will be scaled to fit the page
		pdf.ImageOptions(imgPath, 0, 0, pageWidth, pageHeight, false, gofpdf.ImageOptions{
			ImageType: "JPEG",
			ReadDpi:   true,
		}, 0, "")

		// Add invisible text layer if we have text for this page
		if i < len(texts) && texts[i] != "" {
			fmt.Printf("Adding %d characters of text to page %d\n", len(texts[i]), i+1)

			// Set text rendering mode to invisible (mode 3)
			// In gofpdf, we can't directly set rendering mode to 3,
			// but we can add white text on white background
			pdf.SetTextColor(255, 255, 255) // White text
			pdf.SetFont("Arial", "", 1)     // Very small font
			pdf.SetXY(0, 0)

			// Add the text as a multi-cell which makes it selectable
			// Split into smaller chunks to avoid issues
			lines := strings.Split(texts[i], "\n")
			yPos := 0.0
			for _, line := range lines {
				if line != "" && yPos < pageHeight {
					pdf.SetXY(0, yPos)
					// Use MultiCell for better text wrapping
					pdf.MultiCell(pageWidth, 1, line, "", "", false)
					yPos += 1
				}
			}
		} else {
			fmt.Printf("No text available for page %d\n", i+1)
		}
	}

	// Save the PDF
	if err := pdf.OutputFileAndClose(pdfPath); err != nil {
		return fmt.Errorf("failed to create PDF: %v", err)
	}

	fmt.Printf("Created searchable PDF: %s\n", pdfPath)
	return nil
}
