package edubase

import (
	"testing"
)

func TestGetBooks(t *testing.T) {
	// create a playwright.Page instance for testing with authentication
	page, browser, pw, err := setupTestPlaywrightWithLogin()
	if err != nil {
		t.Skipf("Skipping test due to missing credentials or authentication failure: %v", err)
	}
	defer pw.Stop()
	defer browser.Close()
	defer page.Close()

	// create a new LibraryProvider instance
	libraryProvider := NewLibraryProvider(page)

	// call the GetBooks method
	_, err = libraryProvider.GetBooks()
	if err != nil {
		t.Errorf("get books failed: %v", err)
	}

	// check if the books slice is not empty
	if len(libraryProvider.Books) == 0 {
		t.Errorf("no books found")
		return // Exit early to avoid panic
	}

	// check if the first book has an ID
	if libraryProvider.Books[0].Id == 0 {
		t.Errorf("book has no ID")
	}

	// check if the first book has a title
	if libraryProvider.Books[0].Title == "" {
		t.Errorf("book has no title")
	}
}
