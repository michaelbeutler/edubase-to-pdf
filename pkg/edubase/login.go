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

	// wait for navigation after form submission
	if err := l.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("could not wait for navigation: %v", err)
	}

	// additional wait for page elements to stabilize
	time.Sleep(l.verifyLoginDelay)

	// check if we're still on login page (login failed and redirected back)
	loginForm := l.page.Locator("input[name='login']")
	loginFormVisible, _ := loginForm.IsVisible()
	if loginFormVisible {
		return fmt.Errorf("login failed (redirected back to login page)")
	}

	// check if login was successful using multiple selectors for account button
	// Try primary selector first
	accountButton := l.page.Locator("#main-navbar > nav > ul.header-controls-nav.d-flex.mr-4 > li:nth-child(5) > div > div.btn.lookup-dropdown.lookup-dropdown_no-space-between.border-0.w-auto.pl-0 > i.svg-icon-user.users-profile-icon.svg-icon-primary__border.mr-2").First()
	isVisible, err = accountButton.IsVisible()
	
	// If primary selector fails, try fallback selectors
	if err != nil || !isVisible {
		// Try simpler user icon selector
		accountButton = l.page.Locator("i.svg-icon-user").First()
		isVisible, err = accountButton.IsVisible()
		
		if err != nil || !isVisible {
			// Try user profile selector
			accountButton = l.page.Locator(".users-profile-icon").First()
			isVisible, err = accountButton.IsVisible()
			
			if err != nil || !isVisible {
				// Try any account/profile related dropdown
				accountButton = l.page.Locator(".lookup-dropdown").First()
				isVisible, err = accountButton.IsVisible()
			}
		}
	}

	if err != nil || !isVisible {
		return fmt.Errorf("login failed (could not find account button)")
	}

	return nil
}
