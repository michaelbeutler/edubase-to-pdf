package edubase

import (
	"strconv"
	"time"

	"github.com/playwright-community/playwright-go"
)

type LibraryProvider struct {
	page         playwright.Page
	baseURL      string
	Books        []Book
	initialDelay time.Duration
}

func NewLibraryProvider(page playwright.Page) *LibraryProvider {
	return &LibraryProvider{
		page:         page,
		baseURL:      "https://app.edubase.ch",
		Books:        []Book{},
		initialDelay: 500 * time.Millisecond,
	}
}

type Book struct {
	Id    int
	Title string
}

func (l *LibraryProvider) GetBooks() ([]Book, error) {
	// wait for the library page to load
	time.Sleep(l.initialDelay)

	// Wait for the library items to be visible
	libraryLocator := l.page.Locator("#libraryItems > li:not(:first-child)")
	if err := libraryLocator.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000), // 10 second timeout for library to load
	}); err != nil {
		// If no library items are found, the account might have no books
		return []Book{}, nil
	}

	libraryItems, err := libraryLocator.All()
	if err != nil {
		return []Book{}, err
	}

	for _, libraryItem := range libraryItems {
		bookId, err := libraryItem.GetAttribute("data-last-available-version")
		if err != nil {
			continue
		}

		title, err := libraryItem.Locator(".lu-library-item-title").First().InnerText()
		if err != nil {
			continue
		}

		bookIdInt, err := strconv.Atoi(bookId)
		if err != nil {
			continue
		}

		l.Books = append(l.Books, Book{
			Id:    bookIdInt,
			Title: title,
		})
	}

	return l.Books, nil
}
