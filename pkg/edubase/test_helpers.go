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

	// Additional browser launch options for CI environments
	var launchOptions playwright.BrowserTypeLaunchOptions
	if headless {
		// CI-specific options for better stability
		launchOptions = playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(true),
			Args: []string{
				"--no-sandbox",
				"--disable-setuid-sandbox",
				"--disable-dev-shm-usage",
				"--disable-web-security",
				"--disable-features=VizDisplayCompositor",
			},
		}
	} else {
		// Local development options
		launchOptions = playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(false),
		}
	}

	browser, err := pw.Chromium.Launch(launchOptions)
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