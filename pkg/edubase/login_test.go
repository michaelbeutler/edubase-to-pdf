package edubase

import (
	"os"
	"testing"
)

func TestLogin(t *testing.T) {
	// Check if required environment variables are set
	email := os.Getenv("EDUBASE_EMAIL")
	password := os.Getenv("EDUBASE_PASSWORD")
	if email == "" || password == "" {
		t.Skipf("Skipping integration test: EDUBASE_EMAIL and EDUBASE_PASSWORD environment variables must be set. Current values - EDUBASE_EMAIL: %q, EDUBASE_PASSWORD: %q", email, password)
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
}

func TestLoginInvalidCredentials(t *testing.T) {
	// create a playwright.Page instance for testing
	page, browser, pw, err := setupTestPlaywright()
	if err != nil {
		t.Fatalf("Failed to setup playwright for invalid credentials test: %v", err)
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
	err = loginProvider.Login(credentials, false)
	if err == nil {
		t.Errorf("login with invalid credentials should have failed")
	}
}
