package cmd

import (
	"os"
	"strconv"
	"testing"

	"github.com/michaelbeutler/edubase-to-pdf/pkg/edubase"
	"github.com/playwright-community/playwright-go"
)

// newTestImportProcess creates a new import process with headless mode set based on CI environment
func newTestImportProcess() *importProcess {
	pw, _ := playwright.Run()
	
	// Use headless mode in CI environment
	headless := os.Getenv("CI") == "true"
	
	browser, _ := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
		Timeout:  playwright.Float(float64(timeout.Milliseconds())),
	})
	
	page, _ := browser.NewPage(playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{
			Width:  *playwright.Int(width),
			Height: *playwright.Int(height),
		},
	})

	loginProvider := edubase.NewLoginProvider(page)
	libraryProvider := edubase.NewLibraryProvider(page)

	return &importProcess{
		page:            page,
		browser:         browser,
		pw:              pw,
		loginProvider:   loginProvider,
		libraryProvider: libraryProvider,
	}
}

func TestImport(t *testing.T) {
	// Check if required environment variables are set
	email := os.Getenv("EDUBASE_EMAIL")
	password := os.Getenv("EDUBASE_PASSWORD")
	bookIdStr := os.Getenv("EDUBASE_BOOK_ID")
	
	if email == "" || password == "" {
		t.Fatalf("Integration test failed: EDUBASE_EMAIL and EDUBASE_PASSWORD environment variables must be set. Current values - EDUBASE_EMAIL: %q, EDUBASE_PASSWORD: %q", email, password)
	}

	// Use default book ID if not provided (same as other tests)
	if bookIdStr == "" {
		bookIdStr = "58216"
		t.Logf("EDUBASE_BOOK_ID not set, using default book ID: %s", bookIdStr)
	}

	if err := playwright.Install(); err != nil {
		t.Fatalf("could not install Playwright: %v", err)
	}

	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		t.Fatalf("could not parse book id: %v", err)
	}

	credentials := edubase.Credentials{
		Email:    email,
		Password: password,
	}

	importProcess := newTestImportProcess()
	importProcess.login(credentials)
	importProcess.bookProvider = edubase.NewBookProvider(importProcess.page, bookId)

	if err := importProcess.bookProvider.Open(1); err != nil {
		t.Fatalf("could not open book: %v", err)
	}

	totalPages, err := importProcess.bookProvider.GetTotalPages()
	if err != nil {
		t.Fatalf("could not get total pages: %v", err)
	}

	if totalPages == 0 {
		t.Fatalf("total pages is 0")
	}

	if err = importProcess.browser.Close(); err != nil {
		t.Fatalf("could not close browser: %v", err)
	}

	if err = importProcess.pw.Stop(); err != nil {
		t.Fatalf("could not stop Playwright: %v", err)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal_filename", "normal_filename"},
		{"file/with/slash", "file_with_slash"},
		{"file\\with\\backslash", "file_with_backslash"},
		{"file:with:colon", "file_with_colon"},
		{"file*with*asterisk", "file_with_asterisk"},
		{"file?with?question", "file_with_question"},
		{"file\"with\"quote", "file_with_quote"},
		{"file<with<less", "file_with_less"},
		{"file>with>greater", "file_with_greater"},
		{"file|with|pipe", "file_with_pipe"},
		{"all/\\:*?\"<>|chars", "all_________chars"},
		{"", ""},
	}

	for _, tt := range tests {
		result := sanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}
