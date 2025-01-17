package cmd

import (
	"github.com/playwright-community/playwright-go"
	"testing"
)

func TestImport(t *testing.T) {
	if err := playwright.Install(); err != nil {
		t.Fatalf("could not install Playwright: %v", err)
	}

	bookId, ok := os.LookupEnv("EDUBASE_BOOK_ID").(int)
	if !ok {
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

	if err = importProcess.browser.Close(); err != nil {
		t.Fatalf("could not close browser: %v", err)
	}

	if err = importProcess.pw.Stop(); err != nil {
		t.Fatalf("could not stop Playwright: %v", err)
	}
}
