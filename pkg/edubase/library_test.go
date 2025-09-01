package edubase

import (
	"os"
	"testing"

	"github.com/playwright-community/playwright-go"
)

func TestGetBooks(t *testing.T) {
	// create a playwright.Page instance for testing
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("Failed to create playwright instance: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		t.Fatalf("Failed to launch browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("Failed to create new page: %v", err)
	}
	defer page.Close()

	// create a new LoginProvider instance
	loginProvider := NewLoginProvider(page)

	// set up test credentials
	credentials := Credentials{
		Email:    os.Getenv("EDUBASE_EMAIL"),
		Password: os.Getenv("EDUBASE_PASSWORD"),
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
