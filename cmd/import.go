package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/michaelbeutler/edubase-to-pdf/pkg/edubase"
	pdfcpu "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/playwright-community/playwright-go"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var screenshotDir string = "screenshots"
var email string = ""
var password string = ""
var maxPages int = -1
var startPage int = 1
var debug bool = false
var imgOverwrite bool = false
var width int = 2560
var height int = 1440
var pageDelay time.Duration = 500 * time.Millisecond
var timeout time.Duration = 5 * time.Minute

func init() {
	importCmd.Flags().StringVarP(&screenshotDir, "temp", "t", "screenshots", "Temporary directory for screenshots these will be used to generate the pdf.")
	importCmd.Flags().StringVarP(&email, "email", "e", "", "Edubase email for login.")
	importCmd.Flags().StringVarP(&password, "password", "p", "", "Edubase password for login.")
	importCmd.Flags().IntVarP(&maxPages, "max-pages", "m", -1, "Max pages to import from the book.")
	importCmd.Flags().IntVarP(&startPage, "start-page", "s", 1, "Start page to import from the book.")
	importCmd.Flags().BoolVarP(&imgOverwrite, "img-overwrite", "o", false, "Overwrite existing screenshots.")
	importCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode. Show browser window.")
	importCmd.Flags().IntVarP(&height, "height", "H", height, "Browser height in pixels this can affect the screenshot quality.")
	importCmd.Flags().IntVarP(&width, "width", "W", width, "Browser width in pixels this can affect the screenshot quality.")
	importCmd.Flags().DurationVarP(&pageDelay, "page-delay", "D", pageDelay, "Delay between pages in milliseconds. This is required to give the browser time to load the page.")
	importCmd.Flags().DurationVarP(&timeout, "timeout", "T", timeout, "Maximum time the app can take to download all pages. (increase this value for large books)")

	importCmd.MarkFlagsRequiredTogether("email", "password")

	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use: "import",
	Long: `Description:
  The import command will sign in to Edubase, fetch the books, and take screenshots of the pages. 
  Screenshots will be used to generate a PDF. The PDF will be saved in the current directory.

Example:
  edubase-to-pdf import -e your_email@example.com -p your_password -s 2 -m 10

  This example will sign in to Edubase using the provided email and password. 
  It will start importing from page 2 and import a maximum of 10 pages. 
  The resulting PDF will be saved in the current directory.

Contact:
  For any issues or questions, please open an issue on the GitHub repository:
  https://github.com/michaelbeutler/edubase-to-pdf/issues`,
	Run: func(cmd *cobra.Command, args []string) {
		err := playwright.Install()
		if err != nil {
			log.Fatalf("could not install Playwright: %v", err)
		}

		importProcess := newImportProcess()

		credentials := edubase.Credentials{
			Email:    email,
			Password: password,
		}

		// if email or password is empty, get credentials from form
		if email == "" || password == "" {
			c, err := edubase.GetCredentials()
			if err != nil {
				log.Fatalf("could not get credentials: %v", err)
			}

			credentials = c
		}

		// login
		importProcess.login(credentials)

		// get books
		books, err := importProcess.getBooks()
		if err != nil {
			log.Fatalf("could not get books: %v", err)
		}

		book := edubase.Book{}
		booksForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[edubase.Book]().Title("Book").OptionsFunc(func() []huh.Option[edubase.Book] {
					return huh.NewOptions(books...)
				}, &books).Key("Title").Value(&book),
			),
		)

		err = booksForm.Run()
		if err != nil {
			log.Fatalf("could not get book id: %v", err)
		}

		// open book
		importProcess.bookProvider = edubase.NewBookProvider(importProcess.page, book.Id)

		err = importProcess.bookProvider.Open(startPage)
		if err != nil {
			log.Fatalf("could not open book: %v", err)
		}

		totalPages, err := importProcess.bookProvider.GetTotalPages()
		if err != nil {
			log.Fatalf("could not get total pages: %v", err)
		}

		if maxPages != -1 {
			totalPages = maxPages
		}

		createDirIfNotExists(screenshotDir)

		barDownloadImg := progressbar.Default(int64(totalPages), "Downloading pages...")
		for i := startPage; i <= (startPage-1)+totalPages; i++ {

			filename := fmt.Sprintf("%s/%d_%d.jpeg", screenshotDir, book.Id, i)

			if _, err := os.Stat(filename); err == nil && !imgOverwrite {
				// file exists, skip screenshot
			} else {
				// wait for page to load
				time.Sleep(pageDelay)
				// take screenshot
				err = importProcess.bookProvider.Screenshot(filename)
				if err != nil {
					log.Fatalf("could not take screenshot: %v", err)
				}
			}

			// next page
			err = importProcess.bookProvider.NextPage()
			if err != nil {
				log.Fatalf("could not navigate to next page: %v", err)
			}

			barDownloadImg.Add(1)
		}

		// Generate PDF from screenshots that are previously taken
		barImgtoPdf := progressbar.Default(int64(totalPages), "Generating PDF...")
		pdfPath := fmt.Sprintf("%s.pdf", sanitizeFilename(book.Title))
		for i := startPage; i <= (startPage-1)+totalPages; i++ {
			filename := fmt.Sprintf("%s/%d_%d.jpeg", screenshotDir, book.Id, i)
			// Generate PDF and append to the sanitized output file
			if err := pdfcpu.ImportImagesFile([]string{filename}, pdfPath, nil, model.NewDefaultConfiguration()); err != nil {
				log.Fatalf("could not import image %s: %v", filename, err)
			}
			barImgtoPdf.Add(1)
		}

		// Read the PDF Syntax
		pdfReadCtx, err := pdfcpu.ReadContextFile(pdfPath)
		if err != nil {
			log.Fatalf("❌ Failed to read PDF file '%s' to validate: %v", pdfPath, err)
		}
		// Validate the number of pages in the PDF
		actualPageCountInPdf := pdfReadCtx.PageCount
		if actualPageCountInPdf < totalPages {
			log.Fatalf("❌ Failed to import all pages! Ebook Pages: %d | Pages in PDF: %d. Maybe delete PDF and try again.", totalPages, actualPageCountInPdf)
		}

		if actualPageCountInPdf > totalPages {
			log.Fatalf("❌ PDF has too many pages! Ebook Pages: %d | Pages in PDF: %d. Maybe delete PDF and try again.", totalPages, actualPageCountInPdf)
		}

		if err = importProcess.browser.Close(); err != nil {
			log.Fatalf("could not close browser: %v", err)
		}
		if err = importProcess.pw.Stop(); err != nil {
			log.Fatalf("could not stop Playwright: %v", err)
		}

	},
}

func sanitizeFilename(filename string) string {
	sanitized := filename
	for _, char := range []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"} {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}
	return sanitized
}

type importProcess struct {
	page            playwright.Page
	browser         playwright.Browser
	pw              *playwright.Playwright
	loginProvider   *edubase.LoginProvider
	bookProvider    *edubase.BookProvider
	libraryProvider *edubase.LibraryProvider
}

func newPlaywrightPage() (playwright.Page, playwright.Browser, *playwright.Playwright) {
	pw, _ := playwright.Run()
	browser, _ := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(!debug),
		Timeout:  playwright.Float(float64(timeout.Milliseconds())),
	})
	page, _ := browser.NewPage(playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{
			Width:  *playwright.Int(width),
			Height: *playwright.Int(height),
		},
	})
	return page, browser, pw
}

func newImportProcess() *importProcess {
	page, browser, pw := newPlaywrightPage()

	loginProvider := edubase.NewLoginProvider(page)
	libraryProvider := edubase.NewLibraryProvider(page)

	return &importProcess{
		page:            page,
		browser:         browser,
		pw:              pw,
		loginProvider:   loginProvider,
		libraryProvider: libraryProvider,
	}
}

func (i *importProcess) login(credentials edubase.Credentials) {
	err := spinner.New().Title("logging in...").
		Action(func() {
			err := i.loginProvider.Login(credentials)
			if err != nil {
				log.Fatalf("could not login: %v", err)
			}

		}).
		Run()

	if err != nil {
		log.Fatalf("could not login: %v", err)
	}
}

func (i *importProcess) getBooks() ([]edubase.Book, error) {
	err := spinner.New().Title("fetching books...").
		Action(func() {
			_, err := i.libraryProvider.GetBooks()
			if err != nil {
				log.Fatalf("could not get books: %v", err)
			}
		}).
		Run()

	return i.libraryProvider.Books, err
}

func createDirIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}
}
