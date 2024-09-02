package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	pdfcpu "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/playwright-community/playwright-go"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var book Book
var screenshotDir string = "screenshots"
var email string = ""
var password string = ""
var maxPages int = -1
var startPage int = 1
var debug bool = false
var signInSleep time.Duration = 2 * time.Second
var width int = 2560
var height int = 1440

func init() {
	importCmd.Flags().StringVarP(&screenshotDir, "temp", "t", "screenshots", "Temporary directory for screenshots.")
	importCmd.Flags().StringVarP(&email, "email", "e", "", "Edubase email.")
	importCmd.Flags().StringVarP(&password, "password", "p", "", "Edubase password.")
	importCmd.Flags().IntVarP(&maxPages, "max-pages", "m", -1, "Max pages to import.")
	importCmd.Flags().IntVarP(&startPage, "start-page", "s", 1, "Start page to import.")
	importCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode. Show browser window.")
	importCmd.Flags().DurationP("sleep", "l", signInSleep, "Sleep duration after login.")
	importCmd.Flags().IntVarP(&height, "height", "H", height, "Browser height.")
	importCmd.Flags().IntVarP(&width, "width", "W", width, "Browser width.")

	importCmd.MarkFlagRequired("email")

	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:   "import [bookId]",
	Short: "Import a book from Edubase",
	Long:  `Import a book from Edubase`,
	Run: func(cmd *cobra.Command, args []string) {
		getLoginForm()
		createDirIfNotExists(screenshotDir)
		importEdubaseBook()
	},
}

func importEdubaseBook() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(!debug),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{
			Width:  height,
			Height: width,
		},
	})
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	err = spinner.New().Title("signing in...").
		Action(func() {
			signIn(page)
		}).
		Run()

	var books []Book
	err = spinner.New().Title("fetching books...").
		Action(func() {
			books, err = getBooks(page)
			if err != nil {
				log.Fatalf("could not get books: %v", err)
			}
		}).
		Run()

	book = getBooksForm(books)

	// wait for navigation
	if err = page.WaitForURL("https://app.edubase.ch/#library"); err != nil {
		log.Fatalf("could not wait for navigation: %v", err)
	}

	// navigate to book
	if _, err = page.Goto(fmt.Sprintf("https://app.edubase.ch/#doc/%d/%d", book.Id, startPage), playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	err = spinner.New().Title("waiting for page to load...").
		Action(func() {
			time.Sleep(3 * time.Second)
		}).
		Run()

	// get .doc-page element
	docPage := page.Locator(".lu-page-svg-container").First()

	// get max page number
	totalPages := getTotalPages(page)
	log.Printf("total pages: %d", totalPages)

	if maxPages > 0 && maxPages < totalPages {
		log.Printf("max pages: %d", maxPages)
		totalPages = maxPages
	}

	nextPageButton := page.Locator("[data-action='next-page']").First()
	bar := progressbar.Default(int64(totalPages))
	for i := startPage; i <= (startPage-1)+totalPages; i++ {
		time.Sleep(500 * time.Millisecond)
		filename := fmt.Sprintf("%s/%d_%d.jpeg", screenshotDir, book.Id, i)

		if _, err = docPage.Screenshot(playwright.LocatorScreenshotOptions{
			Path:    playwright.String(filename),
			Quality: playwright.Int(100),
			Type:    playwright.ScreenshotTypeJpeg,
		}); err != nil {
			log.Fatalf("could not create screenshot: %v", err)
		}

		nextPageButton.Click()

		// generate pdf
		pdfcpu.ImportImagesFile([]string{filename}, fmt.Sprintf("%d.pdf", book.Id), nil, model.NewDefaultConfiguration())
		bar.Add(1)
	}

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}

	log.Printf("done")
}

func getLoginForm() {
	if email != "" && password != "" {
		return
	}

	loginForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Email").Value(&email),
			huh.NewInput().Title("Password").Value(&password).EchoMode(huh.EchoModePassword),
		),
	)

	err := loginForm.Run()
	if err != nil {
		log.Fatalf("could not get email: %v", err)
	}
}

func getBooksForm(books []Book) Book {
	book := Book{}
	booksForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[Book]().Title("Book").OptionsFunc(func() []huh.Option[Book] {
				return huh.NewOptions(books...)
			}, &books).Key("Title").Value(&book),
		),
	)

	err := booksForm.Run()
	if err != nil {
		log.Fatalf("could not get book id: %v", err)
	}

	return book
}

type Book struct {
	Id    int
	Title string
}

func getBooks(page playwright.Page) ([]Book, error) {
	var books []Book

	libraryItems, err := page.Locator("#libraryItems > li:not(:first-child)").All()
	if err != nil {
		return []Book{}, fmt.Errorf("could not get books: %v", err)
	}

	for _, libraryItem := range libraryItems {
		bookId, err := libraryItem.GetAttribute("data-last-available-version")
		if err != nil {
			fmt.Errorf("could not get book id: %v", err)
			continue
		}

		title, err := libraryItem.Locator(".lu-library-item-title").First().InnerText()
		if err != nil {
			fmt.Errorf("could not get book title: %v", err)
			continue
		}

		bookIdInt, err := strconv.Atoi(bookId)
		if err != nil {
			fmt.Errorf("could not convert book id: %v", err)
			continue
		}

		books = append(books, Book{
			Id:    bookIdInt,
			Title: title,
		})
	}

	return books, nil
}

func signIn(page playwright.Page) {
	if _, err := page.Goto("https://app.edubase.ch/#promo?popup=login"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	// get login input
	loginInput := page.Locator("input[name='login']")
	loginInput.Fill(email)

	// get password input
	passwordInput := page.Locator("input[name='password']")
	passwordInput.Fill(password)

	// submit form
	if err := page.Locator("button[type='submit']").Click(); err != nil {
		log.Fatalf("could not click: %v", err)
	}

	time.Sleep(2 * time.Second)
}

func getTotalPages(page playwright.Page) int {
	rawTotalPages, err := page.Locator("#pagination > div > span").First().InnerText()
	if err != nil {
		log.Fatalf("could not get max page number: %v", err)
	}

	re := regexp.MustCompile("[0-9]+")
	totalPagesString := re.FindAllString(rawTotalPages, -1)
	totalPages, err := strconv.Atoi(totalPagesString[0])
	if err != nil {
		log.Fatalf("could not convert max page number: %v", err)
	}

	return totalPages
}

func createDirIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}
}

func isValidBookId(bookId int) bool {
	return bookId >= 10000
}
