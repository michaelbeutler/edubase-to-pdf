package edubase

import (
	"fmt"
	"os"

	"github.com/playwright-community/playwright-go"
)

// setupTestPlaywright creates a playwright instance, browser, and page for testing
// It uses headless mode when running in CI
func setupTestPlaywright() (playwright.Page, playwright.Browser, *playwright.Playwright, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to run playwright: %w", err)
	}

	// Use headless mode in CI environment
	headless := os.Getenv("CI") == "true"

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		pw.Stop()
		return nil, nil, nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, nil, nil, fmt.Errorf("failed to create page: %w", err)
	}

	return page, browser, pw, nil
}