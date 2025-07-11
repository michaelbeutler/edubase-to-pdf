package cmd

import (
	"os"
	"strconv"
	"testing"

	"github.com/michaelbeutler/edubase-to-pdf/pkg/edubase"
	"github.com/playwright-community/playwright-go"
)

func TestImport(t *testing.T) {
	if err := playwright.Install(); err != nil {
		t.Fatalf("could not install Playwright: %v", err)
	}

	bookId, err := strconv.Atoi(os.Getenv("EDUBASE_BOOK_ID"))
	if err != nil {
		t.Fatalf("could not get book id from environment")
	}

	credentials := edubase.Credentials{
		Email:    os.Getenv("EDUBASE_EMAIL"),
		Password: os.Getenv("EDUBASE_PASSWORD"),
	}

	importProcess := newImportProcess()
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
