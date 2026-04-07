package edubase

import (
	"fmt"
	"strconv"
	"time"

	"github.com/playwright-community/playwright-go"
)

type LibraryProvider struct {
	page               playwright.Page
	baseURL            string
	Books              []Book
	timeout            time.Duration
	stabilizationDelay time.Duration
}

func NewLibraryProvider(page playwright.Page) *LibraryProvider {
	return &LibraryProvider{
		page:               page,
		baseURL:            "https://app.edubase.ch",
		Books:              []Book{},
		timeout:            15 * time.Second,
		stabilizationDelay: 2 * time.Second,
	}
}

type Book struct {
	Id    int
	Title string
}

func (l *LibraryProvider) GetBooks() ([]Book, error) {
	// wait for at least one library item to be visible in the DOM
	itemLocator := l.page.Locator("#libraryItems > li:not(:first-child)")
	err := itemLocator.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(float64(l.timeout.Milliseconds())),
	})
	if err != nil {
		return []Book{}, fmt.Errorf("timed out waiting for library items to appear: %w", err)
	}

	// wait for network to settle so all books (including paid) finish loading
	_ = l.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(float64(l.timeout.Milliseconds())),
	})

	// allow final DOM mutations after last API response
	time.Sleep(l.stabilizationDelay)

	libraryItems, err := itemLocator.All()
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
