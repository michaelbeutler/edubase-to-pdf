package edubase

import (
	"fmt"
	"time"

	"bufio"
	"os"
	"strings"

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

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return Credentials{}, err
	}
	credentials.Email = strings.TrimSpace(email)

	fmt.Print("Password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		return Credentials{}, err
	}
	credentials.Password = strings.TrimSpace(password)

	return credentials, nil
}

func (l *LoginProvider) Login(credentials Credentials, manualLogin bool) error {
	if err := l.setupLoginPage(); err != nil {
		return err
	}

	if manualLogin {
		return l.handleManualLogin()
	}

	return l.handleAutomaticLogin(credentials)
}

func (l *LoginProvider) setupLoginPage() error {
	// clear all cookies and local storage
	if err := l.page.Context().ClearCookies(); err != nil {
		return fmt.Errorf("could not clear cookies: %v", err)
	}

	// go to login page
	if _, err := l.page.Goto(l.baseURL); err != nil {
		return fmt.Errorf("could not go to base page: %v", err)
	}

	// wait for page to load
	if err := l.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("could not wait for navigation: %v", err)
	}

	// press login button
	if err := l.page.Locator("button[data-open='loginModal']").Click(); err != nil {
		return fmt.Errorf("could not click login button: %v", err)
	}

	return nil
}

func (l *LoginProvider) handleManualLogin() error {
	// wait for user to complete login
	if err := l.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("could not wait for navigation: %v", err)
	}

	return l.waitForLoginSuccess()
}

func (l *LoginProvider) handleAutomaticLogin(credentials Credentials) error {
	if err := l.fillLoginForm(credentials); err != nil {
		return err
	}

	if err := l.submitLoginForm(); err != nil {
		return err
	}

	// wait for login to complete
	time.Sleep(l.verifyLoginDelay)

	return l.verifyLoginSuccess()
}

func (l *LoginProvider) fillLoginForm(credentials Credentials) error {
	// get login input
	loginInput := l.page.Locator("input[name='login']")
	loginInput.Fill(credentials.Email)

	// check if input is visible
	isVisible, err := loginInput.IsVisible()
	if err != nil || !isVisible {
		return fmt.Errorf("could not fill login input: %v", err)
	}

	// wait for password fill delay
	time.Sleep(l.passwordFillDelay)

	// get password input
	passwordInput := l.page.Locator("input[name='password']")
	passwordInput.Fill(credentials.Password)

	return nil
}

func (l *LoginProvider) submitLoginForm() error {
	// submit form
	if err := l.page.Locator("button[type='submit']").Click(); err != nil {
		return fmt.Errorf("could not submit login form: %v", err)
	}

	// wait for login to complete
	if err := l.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("could not wait for navigation: %v", err)
	}

	return nil
}

func (l *LoginProvider) waitForLoginSuccess() error {
	//check in a while loop if login was successful (check for account button)
	for {
		accountButton := l.getAccountButton()
		isVisible, err := accountButton.IsVisible()
		if accountButton != nil && err == nil && isVisible {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (l *LoginProvider) verifyLoginSuccess() error {
	// check if login was successful (check for account button)
	accountButton := l.getAccountButton()
	isVisible, err := accountButton.IsVisible()
	if accountButton == nil || err != nil || !isVisible {
		return fmt.Errorf("login failed (could not find account button)")
	}
	return nil
}

func (l *LoginProvider) getAccountButton() playwright.Locator {
	return l.page.Locator("#main-navbar > nav > ul.header-controls-nav.d-flex.mr-4 > li:nth-child(5) > div > div.btn.lookup-dropdown.lookup-dropdown_no-space-between.border-0.w-auto.pl-0 > i.svg-icon-user.users-profile-icon.svg-icon-primary__border.mr-2").First()
}
