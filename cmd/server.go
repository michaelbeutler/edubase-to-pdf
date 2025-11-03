package cmd

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/michaelbeutler/edubase-to-pdf/pkg/edubase"
	pdfcpu "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cobra"
)

//go:embed web/index.html
var clientHTML []byte

// Sentinel errors
var (
	ErrAuthFailed       = errors.New("authentication failed")
	ErrResponseWritten  = errors.New("response headers already written")
)

const (
	// Server configuration defaults
	defaultServerPort = 8080
	defaultServerHost = "0.0.0.0"

	// HTTP server timeouts
	serverReadTimeout     = 15 * time.Second
	serverWriteTimeout    = 5 * time.Minute
	serverIdleTimeout     = 60 * time.Second
	serverShutdownTimeout = 30 * time.Second

	// Browser configuration
	browserTimeout    = 5 * time.Minute
	browserWidth      = 2560
	browserHeight     = 1440
	screenshotDelay   = 500 * time.Millisecond

	// Temp directory prefix
	tempDirPrefix = "edubase-download-*"
)

var (
	serverPort int
	serverHost string
)

func init() {
	serverCmd.Flags().IntVarP(&serverPort, "port", "P", defaultServerPort, "Port for the HTTP server")
	serverCmd.Flags().StringVarP(&serverHost, "host", "H", defaultServerHost, "Host address for the HTTP server")

	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start HTTP server for PDF downloads",
	Long: `Start an HTTP server that allows clients to download PDFs from Edubase.

The server provides a stateless API endpoint that accepts book download requests
and streams PDF responses. All request parameters are validated and proper error
responses are returned.

Example:
  edubase-to-pdf server --port 8080

API Endpoint:
  POST /download
  
Request Body:
  {
    "email": "your_email@example.com",
    "password": "your_password",
    "book_id": 12345,
    "start_page": 1,
    "max_pages": -1
  }

Response:
  - 200 OK: PDF file stream
  - 400 Bad Request: Invalid request parameters
  - 401 Unauthorized: Authentication failed
  - 500 Internal Server Error: Processing error`,
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure playwright is installed
		if err := playwright.Install(); err != nil {
			log.Fatalf("could not install Playwright: %v", err)
		}

		server := newHTTPServer(serverHost, serverPort)
		server.start()
	},
}

// DownloadRequest represents the request body for PDF download
type DownloadRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	BookID    int    `json:"book_id"`
	StartPage int    `json:"start_page"`
	MaxPages  int    `json:"max_pages"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// httpServer manages the HTTP server
type httpServer struct {
	server *http.Server
}

// newHTTPServer creates a new HTTP server instance
func newHTTPServer(host string, port int) *httpServer {
	mux := http.NewServeMux()
	
	s := &httpServer{
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", host, port),
			Handler:      mux,
			ReadTimeout:  serverReadTimeout,
			WriteTimeout: serverWriteTimeout,
			IdleTimeout:  serverIdleTimeout,
		},
	}

	// Register handlers
	mux.HandleFunc("/", s.handleClient)
	mux.HandleFunc("/download", s.handleDownload)
	mux.HandleFunc("/health", s.handleHealth)

	return s
}

// start begins listening and serving HTTP requests
func (s *httpServer) start() {
	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Starting HTTP server on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

// handleClient serves the web client interface
func (s *httpServer) handleClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed")
		return
	}

	// Only serve the root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(clientHTML)
}

// handleHealth responds to health check requests
func (s *httpServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleDownload processes PDF download requests
func (s *httpServer) handleDownload(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST method is allowed")
		return
	}

	// Parse and validate request
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON request body")
		return
	}

	if err := s.validateRequest(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Process the download request
	if err := s.processDownload(w, &req); err != nil {
		log.Printf("Download processing error: %v", err)
		// Don't write error response if headers were already written
		if errors.Is(err, ErrResponseWritten) {
			return
		}
		if errors.Is(err, ErrAuthFailed) {
			s.writeError(w, http.StatusUnauthorized, "auth_failed", "Authentication failed")
		} else {
			s.writeError(w, http.StatusInternalServerError, "processing_error", "Failed to process request")
		}
	}
}

// validateRequest validates the download request
func (s *httpServer) validateRequest(req *DownloadRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if req.BookID <= 0 {
		return fmt.Errorf("book_id must be a positive integer")
	}
	if req.StartPage <= 0 {
		return fmt.Errorf("start_page must be a positive integer")
	}
	if req.MaxPages == 0 || (req.MaxPages < 0 && req.MaxPages != -1) {
		return fmt.Errorf("max_pages must be -1 (all pages) or a positive integer")
	}
	return nil
}

// processDownload handles the actual PDF generation and streaming
func (s *httpServer) processDownload(w http.ResponseWriter, req *DownloadRequest) error {
	// Create temporary directory for this request
	tempDir, err := s.createTempDir()
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup playwright and browser
	pw, browser, page, err := s.setupBrowser()
	if err != nil {
		return err
	}
	defer s.cleanupBrowser(pw, browser, page)

	// Authenticate user
	if err := s.authenticateUser(page, req); err != nil {
		return fmt.Errorf("authentication failed")
	}

	// Download pages
	totalPages, err := s.downloadPages(page, req, tempDir)
	if err != nil {
		return err
	}

	// Generate and stream PDF
	return s.generateAndStreamPDF(w, req, tempDir, totalPages)
}

// createTempDir creates a temporary directory for screenshots
func (s *httpServer) createTempDir() (string, error) {
	return os.MkdirTemp("", tempDirPrefix)
}

// setupBrowser initializes playwright and creates a browser instance
func (s *httpServer) setupBrowser() (*playwright.Playwright, playwright.Browser, playwright.Page, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to run playwright: %w", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Timeout:  playwright.Float(float64(browserTimeout.Milliseconds())),
	})
	if err != nil {
		pw.Stop()
		return nil, nil, nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{
			Width:  *playwright.Int(browserWidth),
			Height: *playwright.Int(browserHeight),
		},
	})
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, nil, nil, fmt.Errorf("failed to create page: %w", err)
	}

	return pw, browser, page, nil
}

// cleanupBrowser closes browser resources
func (s *httpServer) cleanupBrowser(pw *playwright.Playwright, browser playwright.Browser, page playwright.Page) {
	if page != nil {
		page.Close()
	}
	if browser != nil {
		browser.Close()
	}
	if pw != nil {
		pw.Stop()
	}
}

// authenticateUser logs in to Edubase
func (s *httpServer) authenticateUser(page playwright.Page, req *DownloadRequest) error {
	loginProvider := edubase.NewLoginProvider(page)
	credentials := edubase.Credentials{
		Email:    req.Email,
		Password: req.Password,
	}
	if err := loginProvider.Login(credentials, false); err != nil {
		return fmt.Errorf("%w: %v", ErrAuthFailed, err)
	}
	return nil
}

// downloadPages downloads all pages from the book
func (s *httpServer) downloadPages(page playwright.Page, req *DownloadRequest, tempDir string) (int, error) {
	bookProvider := edubase.NewBookProvider(page, req.BookID)
	if err := bookProvider.Open(req.StartPage); err != nil {
		return 0, fmt.Errorf("failed to open book: %w", err)
	}

	totalPages, err := bookProvider.GetTotalPages()
	if err != nil {
		return 0, fmt.Errorf("failed to get total pages: %w", err)
	}

	pagesToDownload := totalPages
	if req.MaxPages > 0 {
		pagesToDownload = req.MaxPages
	}

	for i := req.StartPage; i <= (req.StartPage-1)+pagesToDownload; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("page_%d.jpeg", i))
		
		time.Sleep(screenshotDelay)
		if err := bookProvider.Screenshot(filename); err != nil {
			return 0, fmt.Errorf("failed to take screenshot: %w", err)
		}

		if err := bookProvider.NextPage(); err != nil {
			return 0, fmt.Errorf("failed to navigate to next page: %w", err)
		}
	}

	return pagesToDownload, nil
}

// generateAndStreamPDF creates PDF from images and streams it to client
func (s *httpServer) generateAndStreamPDF(w http.ResponseWriter, req *DownloadRequest, tempDir string, pagesToDownload int) error {
	pdfPath := filepath.Join(tempDir, "output.pdf")
	
	for i := req.StartPage; i <= (req.StartPage-1)+pagesToDownload; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("page_%d.jpeg", i))
		if err := pdfcpu.ImportImagesFile([]string{filename}, pdfPath, nil, model.NewDefaultConfiguration()); err != nil {
			return fmt.Errorf("failed to generate PDF: %w", err)
		}
	}

	return s.streamPDF(w, pdfPath, req.BookID)
}

// streamPDF streams the PDF file to the client
func (s *httpServer) streamPDF(w http.ResponseWriter, pdfPath string, bookID int) error {
	pdfFile, err := os.Open(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer pdfFile.Close()

	stat, err := pdfFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get PDF stats: %w", err)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=book_%d.pdf", bookID))
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	w.WriteHeader(http.StatusOK)

	// After WriteHeader, response has started - wrap any errors with ErrResponseWritten
	if _, err := io.Copy(w, pdfFile); err != nil {
		return fmt.Errorf("%w: failed to stream PDF: %v", ErrResponseWritten, err)
	}

	return nil
}

// writeError writes an error response
func (s *httpServer) writeError(w http.ResponseWriter, status int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}
