package edubase

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/playwright-community/playwright-go"
)

type LoginProvider struct {
	page              playwright.Page
	baseURL           string
	passwordFillDelay time.Duration
	verifyLoginDelay  time.Duration
}

func NewLoginProvider(page playwright.Page) *LoginProvider {
	return &LoginProvider{
		page:              page,
		baseURL:           "https://app.edubase.ch",
		passwordFillDelay: 500 * time.Millisecond,
		verifyLoginDelay:  500 * time.Millisecond,
	}
}

type Credentials struct {
	Email    string
	Password string
}

func GetCredentials() (Credentials, error) {
	credentials := Credentials{}

	loginForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Email").Value(&credentials.Email),
			huh.NewInput().Title("Password").Value(&credentials.Password).EchoMode(huh.EchoModePassword),
		),
	)

	if err := loginForm.Run(); err != nil {
		return Credentials{}, err
	}

	return credentials, nil
}

func (l *LoginProvider) LoginWithRetry(credentials Credentials, maxRetries int, manualLogin bool) error {
	if manualLogin {
		return l.LoginManually()
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := l.Login(credentials)
		if err == nil {
			return nil
		}
		
		lastErr = err
		if attempt < maxRetries {
			fmt.Printf("Login attempt %d failed (%v), retrying...\n", attempt, err)
			time.Sleep(2 * time.Second) // Wait before retry
		}
	}
	
	return fmt.Errorf("login failed after %d attempts, last error: %v", maxRetries, lastErr)
}

func (l *LoginProvider) LoginManually() error {
	// Navigate to login page
	if _, err := l.page.Goto(fmt.Sprintf("%s/#promo?popup=login", l.baseURL)); err != nil {
		return fmt.Errorf("could not go to login page: %v", err)
	}

	fmt.Println("Please login manually in the browser window...")
	fmt.Println("The application will continue once you have successfully logged in.")
	
	// Wait for the URL to change away from the login page
	timeout := 5 * time.Minute
	startTime := time.Now()
	
	for time.Since(startTime) < timeout {
		currentURL := l.page.URL()
		
		// Check if we're no longer on the login page
		if !strings.Contains(currentURL, "popup=login") && !strings.Contains(currentURL, "#promo") {
			// Additional check to ensure we're actually logged in
			time.Sleep(l.verifyLoginDelay)
			accountButton := l.page.Locator("#main-navbar > nav > ul.header-controls-nav.d-flex.mr-4 > li:nth-child(5) > div > div.btn.lookup-dropdown.lookup-dropdown_no-space-between.border-0.w-auto.pl-0 > i.svg-icon-user.users-profile-icon.svg-icon-primary__border.mr-2").First()
			isVisible, err := accountButton.IsVisible()
			if err == nil && isVisible {
				fmt.Println("Login successful!")
				return nil
			}
		}
		
		time.Sleep(1 * time.Second)
	}
	
	return fmt.Errorf("manual login timeout: no successful login detected within %v", timeout)
}

func (l *LoginProvider) Login(credentials Credentials) error {
	// go to login page with increased timeout and handle potential reloads
	timeout := 60 * time.Second
	if _, err := l.page.Goto(fmt.Sprintf("%s/#promo?popup=login", l.baseURL), playwright.PageGotoOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	}); err != nil {
		return fmt.Errorf("could not go to login page: %v", err)
	}

	// Wait for the page to stabilize (handle potential reloads)
	time.Sleep(2 * time.Second)
	
	// Wait for login form to be ready
	loginInput := l.page.Locator("input[name='login']")
	if err := loginInput.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(30000), // 30 seconds
	}); err != nil {
		return fmt.Errorf("login form not ready: %v", err)
	}

	// Fill login credentials
	if err := loginInput.Fill(credentials.Email); err != nil {
		return fmt.Errorf("could not fill email: %v", err)
	}

	// check if input is visible
	isVisible, err := loginInput.IsVisible()
	if err != nil || !isVisible {
		return fmt.Errorf("could not fill login input: %v", err)
	}

	// wait for password fill delay
	time.Sleep(l.passwordFillDelay)

	// get password input
	passwordInput := l.page.Locator("input[name='password']")
	if err := passwordInput.Fill(credentials.Password); err != nil {
		return fmt.Errorf("could not fill password: %v", err)
	}

	// submit form
	if err := l.page.Locator("button[type='submit']").Click(); err != nil {
		return fmt.Errorf("could not submit login form: %v", err)
	}

	// wait for login to complete with longer timeout
	if err := l.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(60000), // 60 seconds
	}); err != nil {
		return fmt.Errorf("could not wait for navigation: %v", err)
	}

	// wait for login to complete
	time.Sleep(l.verifyLoginDelay)

	// check if login was successful (check for account button)
	accountButton := l.page.Locator("#main-navbar > nav > ul.header-controls-nav.d-flex.mr-4 > li:nth-child(5) > div > div.btn.lookup-dropdown.lookup-dropdown_no-space-between.border-0.w-auto.pl-0 > i.svg-icon-user.users-profile-icon.svg-icon-primary__border.mr-2").First()
	isVisible, err = accountButton.IsVisible()
	if accountButton == nil || err != nil || !isVisible {
		return fmt.Errorf("login failed (could not find account button)")
	}

	return nil
}
