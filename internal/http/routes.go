package http

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed static
var staticFiles embed.FS

// SetupRoutes configures all HTTP routes for the server
func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Printf("Warning: Could not load static files: %v", err)
	} else {
		mux.Handle("/", http.FileServer(http.FS(staticFS)))
	}

	// API endpoints
	mux.HandleFunc("/api/books", s.GetBooksHandler)
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.LoginHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.StartDownloadHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/download/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/download/")

		if strings.HasSuffix(path, "/pdf") {
			s.DownloadPDFHandler(w, r)
		} else if strings.HasSuffix(path, "/events") {
			s.SSEHandler(w, r)
		} else {
			// Get status
			s.GetDownloadStatusHandler(w, r)
		}
	})

	// Create session endpoint
	mux.HandleFunc("/api/session", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sessionID := s.CreateSession()
		// Set Secure flag only when the request was received over TLS.
		// This ensures cookies are marked Secure in production (HTTPS)
		// while still allowing local development over HTTP (e.g., localhost).
		secureFlag := false
		if r.TLS != nil {
			secureFlag = true
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   secureFlag,
			SameSite: http.SameSiteStrictMode,
		})

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"session_id":"` + sessionID + `"}`))
	})

	return mux
}

// Start starts the HTTP server on the specified address
func (s *Server) Start(addr string) error {
	mux := s.SetupRoutes()
	log.Printf("Starting HTTP server on %s", addr)
	log.Printf("Web UI available at http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}
