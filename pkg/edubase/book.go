package edubase

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/playwright-community/playwright-go"
)

type BookProvider struct {
	page         playwright.Page
	baseURL      string
	bookId       int
	initialDelay time.Duration
}

func NewBookProvider(page playwright.Page, id int) *BookProvider {
	return &BookProvider{
		page:         page,
		baseURL:      "https://app.edubase.ch",
		bookId:       id,
		initialDelay: 500 * time.Millisecond,
	}
}

func (b *BookProvider) Open(page int) error {
	time.Sleep(b.initialDelay)

	// navigate to book
	if _, err := b.page.Goto(fmt.Sprintf("%s/#doc/%d/%d", b.baseURL, b.bookId, page), playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return fmt.Errorf("could not open book: %v", err)
	}

	return nil
}

func (b *BookProvider) GetTotalPages() (int, error) {
	time.Sleep(b.initialDelay)

	// Wait for the pagination element to be visible and contain numbers
	paginationLocator := b.page.Locator("#pagination > div > span").First()

	// Wait for the element to be visible with a timeout
	if err := paginationLocator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000), // 10 second timeout
	}); err != nil {
		return 0, fmt.Errorf("pagination element not found or not visible: %v", err)
	}

	// Retry getting the text until it contains numbers (max 10 attempts)
	var rawTotalPages string
	var err error
	re := regexp.MustCompile("[0-9]+")

	for i := 0; i < 10; i++ {
		rawTotalPages, err = paginationLocator.InnerText()
		if err != nil {
			return 0, fmt.Errorf("could not get max page number: %v", err)
		}

		// Check if we have numbers
		if re.MatchString(rawTotalPages) {
			break
		}

		// Wait a bit before retrying
		time.Sleep(500 * time.Millisecond)
	}

	totalPagesString := re.FindAllString(rawTotalPages, -1)

	if len(totalPagesString) == 0 {
		return 0, fmt.Errorf("could not find max page number in pagination text: %q (element loaded but content not populated after retries)", rawTotalPages)
	}

	totalPages, err := strconv.Atoi(totalPagesString[0])
	if err != nil {
		return 0, fmt.Errorf("could not convert max page number: %v", err)
	}

	return totalPages, nil
}

func (b *BookProvider) NextPage() error {
	// navigate to next page
	nextPageButton := b.page.Locator("[data-action='next-page']").First()

	if err := nextPageButton.Click(); err != nil {
		return fmt.Errorf("could not click next page button: %v", err)
	}

	return nil
}

func (b *BookProvider) Screenshot(filename string) error {
	// check if filename is empty
	if filename == "" {
		return fmt.Errorf("filename is empty")
	}

	// check if filename has the correct extension
	if !regexp.MustCompile(`.*\.jpe?g`).MatchString(filename) {
		return fmt.Errorf("filename has the wrong extension")
	}

	// get .doc-page element
	docPage := b.page.Locator(".lu-page-svg-container").First()

	// take screenshot
	if _, err := docPage.Screenshot(playwright.LocatorScreenshotOptions{
		Path:    playwright.String(filename),
		Quality: playwright.Int(100),
		Type:    playwright.ScreenshotTypeJpeg,
	}); err != nil {
		return fmt.Errorf("could not create screenshot: %v", err)
	}

	return nil
}

// GetPageText extracts all visible text from the current page
func (b *BookProvider) GetPageText() (string, error) {
	// Wait a bit for the page to fully load
	time.Sleep(500 * time.Millisecond)

	// Use JavaScript to get text ONLY from the page SVG, excluding navigation
	allText, err := b.page.EvaluateHandle(`() => {
		// Get only text from the current page SVG, not navigation or UI
		const pageContainer = document.querySelector('.lu-page-svg-container svg, .lu-page svg');
		if (!pageContainer) {
			console.log('No SVG container found');
			return '';
		}
		
		// Get all text elements within the SVG
		const textElements = pageContainer.querySelectorAll('text, tspan');
		let text = '';
		const seenTexts = new Set(); // Avoid duplicates
		
		textElements.forEach(el => {
			const content = (el.textContent || el.innerText || '').trim();
			if (content && !seenTexts.has(content)) {
				seenTexts.add(content);
				text += content + ' ';
			}
		});
		
		console.log('Extracted from SVG:', text.substring(0, 100), 'Total:', text.length);
		return text.trim();
	}`)

	if err == nil {
		text, err := allText.JSONValue()
		if err == nil {
			if str, ok := text.(string); ok && len(str) > 0 {
				fmt.Printf("Successfully extracted text from page SVG (%d chars) - Preview: %.80s...\n", len(str), str)
				return str, nil
			}
		}
	}

	fmt.Printf("Warning: No text could be extracted from page SVG\n")
	return "", nil // Return empty string instead of error so PDF generation continues
}
