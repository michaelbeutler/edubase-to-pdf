package edubase

import (
	"os"
	"testing"
)

func TestGetBooks(t *testing.T) {
	// Check if required environment variables are set
	email := os.Getenv("EDUBASE_EMAIL")
	password := os.Getenv("EDUBASE_PASSWORD")
	if email == "" || password == "" {
		t.Fatalf("Integration test failed: EDUBASE_EMAIL and EDUBASE_PASSWORD environment variables must be set. Current values - EDUBASE_EMAIL: %q, EDUBASE_PASSWORD: %q", email, password)
	}

	// create a playwright.Page instance for testing
	page, browser, pw, err := setupTestPlaywright()
	if err != nil {
		t.Fatalf("Failed to setup playwright: %v", err)
	}
	defer pw.Stop()
	defer browser.Close()
	defer page.Close()

	// create a new LoginProvider instance
	loginProvider := NewLoginProvider(page)

	// set up test credentials
	credentials := Credentials{
		Email:    email,
		Password: password,
	}
	manualLogin := false

	// call the Login method with the test credentials
	err = loginProvider.Login(credentials, manualLogin)
	if err != nil {
		t.Errorf("login failed: %v", err)
	}

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
		return  // Exit early to avoid panic
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
