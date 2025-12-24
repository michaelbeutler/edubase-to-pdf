package cmd

import (
	"fmt"
	"log"

	"github.com/michaelbeutler/edubase-to-pdf/internal/http"
	"github.com/spf13/cobra"
)

var (
	serverPort string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server for edubase book downloads",
	Long: `Start an HTTP server that provides API endpoints for:
- Listing books from edubase library
- Starting book downloads
- Monitoring download progress
- Downloading completed PDF files
- Real-time progress updates via Server-Sent Events (SSE)

The server uses session-based authentication to maintain state across requests.

Example usage:
  edubase-to-pdf server --port 8080

Then access the API at http://localhost:8080/api/`,
	Run: func(cmd *cobra.Command, args []string) {
		server := http.NewServer()
		addr := fmt.Sprintf(":%s", serverPort)

		log.Printf("Starting edubase HTTP server on port %s", serverPort)
		log.Println("API endpoints:")
		log.Println("  POST /api/session - Create a new session")
		log.Println("  POST /api/login - Login to edubase")
		log.Println("  GET  /api/books - List available books")
		log.Println("  POST /api/download - Start a book download")
		log.Println("  GET  /api/download/:jobId - Get download status")
		log.Println("  GET  /api/download/:jobId/pdf - Download completed PDF")
		log.Println("  GET  /api/download/:jobId/events - SSE progress stream")

		if err := server.Start(addr); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVarP(&serverPort, "port", "p", "8080", "Port to run the server on")
}
