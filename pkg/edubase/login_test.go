package edubase

import (
	"os"
	"testing"
)

func TestLogin(t *testing.T) {
	// Skip integration test if required environment variables are not set
	if shouldSkipIntegrationTest() {
		t.Skip("Skipping integration test: EDUBASE_EMAIL and EDUBASE_PASSWORD environment variables must be set")
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
		Email:    os.Getenv("EDUBASE_EMAIL"),
		Password: os.Getenv("EDUBASE_PASSWORD"),
	}

	// call the Login method with the test credentials
	err = loginProvider.Login(credentials)
	if err != nil {
		t.Errorf("login failed: %v", err)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
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
		Email:    "foo",
		Password: "bar",
	}

	// call the Login method with the test credentials
	err = loginProvider.Login(credentials)
	if err == nil {
		t.Errorf("login with invalid credentials should have failed")
	}
}
