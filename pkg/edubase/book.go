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

	rawTotalPages, err := b.page.Locator("#pagination > div > span").First().InnerText()
	if err != nil {
		return 0, fmt.Errorf("could not get max page number: %v", err)
	}

	re := regexp.MustCompile("[0-9]+")
	totalPagesString := re.FindAllString(rawTotalPages, -1)

	if len(totalPagesString) == 0 {
		return 0, fmt.Errorf("could not find max page number: %s", rawTotalPages)
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
