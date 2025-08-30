package edubase

import (
	"os"

	"github.com/playwright-community/playwright-go"
)

// setupTestPlaywright creates a playwright instance, browser, and page for testing
// It uses headless mode when running in CI
func setupTestPlaywright() (playwright.Page, playwright.Browser, *playwright.Playwright, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, nil, nil, err
	}

	// Use headless mode in CI environment
	headless := os.Getenv("CI") == "true"

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		pw.Stop()
		return nil, nil, nil, err
	}

	page, err := browser.NewPage()
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, nil, nil, err
	}

	return page, browser, pw, nil
}

// shouldSkipIntegrationTest checks if integration test should be skipped
// due to missing environment variables
func shouldSkipIntegrationTest() bool {
	email := os.Getenv("EDUBASE_EMAIL")
	password := os.Getenv("EDUBASE_PASSWORD")
	return email == "" || password == ""
}