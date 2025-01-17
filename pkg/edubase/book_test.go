package edubase

import (
	"os"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

func TestOpenBookAtPage1(t *testing.T) {
	// create a playwright.Page instance for testing
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("failed to create playwright instance: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		t.Fatalf("failed to launch browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("failed to create new page: %v", err)
	}
	defer page.Close()

	// create a new BookProvider instance
	bookProvider := NewBookProvider(page, 58216)

	// call the Open method
	err = bookProvider.Open(1)

	if err != nil {
		t.Errorf("failed to open book: %v", err)
	}

	// wait for the page to load
	time.Sleep(2 * time.Second)

	// check if the page has been opened
	title, err := page.Title()
	if err != nil {
		t.Errorf("could not get page title: %v", err)
	}

	if title != "Edubase Reader" {
		t.Errorf("unexpected page title: %s", title)
	}
}

func TestGetTotalPages(t *testing.T) {
	// create a playwright.Page instance for testing
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("failed to create playwright instance: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		t.Fatalf("failed to launch browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("failed to create new page: %v", err)
	}
	defer page.Close()

	// create a new BookProvider instance
	bookProvider := NewBookProvider(page, 58216)

	// call the Open method
	err = bookProvider.Open(1)

	if err != nil {
		t.Errorf("failed to open book: %v", err)
	}

	// wait for the page to load
	time.Sleep(2 * time.Second)

	// call the GetTotalPages method
	totalPages, err := bookProvider.GetTotalPages()
	if err != nil {
		t.Errorf("get total pages failed: %v", err)
	}

	// check if the total number of pages is greater than 0
	if totalPages != 38 {
		t.Errorf("unexpected total number of pages: %d", totalPages)
	}
}

func TestNextPage(t *testing.T) {
	// create a playwright.Page instance for testing
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("failed to create playwright instance: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		t.Fatalf("failed to launch browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("failed to create new page: %v", err)
	}
	defer page.Close()

	// create a new BookProvider instance
	bookProvider := NewBookProvider(page, 58216)

	// call the Open method
	err = bookProvider.Open(1)

	if err != nil {
		t.Errorf("open book failed: %v", err)
	}

	// wait for the page to load
	time.Sleep(2 * time.Second)

	// call the NextPage method
	err = bookProvider.NextPage()
	if err != nil {
		t.Errorf("failed to get next page: %v", err)
	}

	// wait for the page to load
	time.Sleep(2 * time.Second)

	// get the current page number
	currentPageInput, err := page.Locator("#pagination > div > div > span").First().InnerText()
	if err != nil {
		t.Errorf("failed to get current page number: %v", err)
	}

	// check if the current page number is 2
	if currentPageInput != "2" {
		t.Errorf("unexpected current page number: %s", currentPageInput)
	}
}
func TestScreenshot(t *testing.T) {
	// create a playwright.Page instance for testing
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("failed to create playwright instance: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		t.Fatalf("failed to launch browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("failed to create new page: %v", err)
	}
	defer page.Close()

	// create a new BookProvider instance
	bookProvider := NewBookProvider(page, 58216)

	// call the Open method
	err = bookProvider.Open(1)
	if err != nil {
		t.Errorf("failed to open book: %v", err)
	}

	// wait for the page to load
	time.Sleep(2 * time.Second)

	// call the Screenshot method
	err = bookProvider.Screenshot("test.jpg")
	if err != nil {
		t.Errorf("screenshot failed: %v", err)
	}

	// check if the screenshot file exists
	if _, err := os.Stat("test.jpg"); os.IsNotExist(err) {
		t.Errorf("screenshot file does not exist")
	}

	// delete the screenshot file
	defer func() {
		err := os.Remove("test.jpg")
		if err != nil {
			t.Errorf("could not delete screenshot")
		}
	}()
}
