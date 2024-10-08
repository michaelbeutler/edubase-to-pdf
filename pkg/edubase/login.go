package edubase

import (
	"fmt"
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

func (l *LoginProvider) Login(credentials Credentials) error {
	// go to login page
	if _, err := l.page.Goto(fmt.Sprintf("%s/#promo?popup=login", l.baseURL)); err != nil {
		return fmt.Errorf("could not go to login page: %v", err)
	}

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
