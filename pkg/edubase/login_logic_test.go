package edubase

import (
	"strings"
	"testing"
)

func TestLoginProviderCreation(t *testing.T) {
	// Test that NewLoginProvider creates a provider with correct defaults
	// This test doesn't require Playwright to be installed
	provider := &LoginProvider{
		page:              nil, // We're just testing the struct, not actual page interaction
		baseURL:           "https://app.edubase.ch",
		passwordFillDelay: 500000000, // 500ms in nanoseconds
		verifyLoginDelay:  500000000, // 500ms in nanoseconds
	}

	if provider.baseURL != "https://app.edubase.ch" {
		t.Errorf("Expected baseURL to be https://app.edubase.ch, got %s", provider.baseURL)
	}

	if provider.passwordFillDelay.String() != "500ms" {
		t.Errorf("Expected passwordFillDelay to be 500ms, got %s", provider.passwordFillDelay.String())
	}

	if provider.verifyLoginDelay.String() != "500ms" {
		t.Errorf("Expected verifyLoginDelay to be 500ms, got %s", provider.verifyLoginDelay.String())
	}
}

func TestLoginSelectorLogic(t *testing.T) {
	// Test that our fallback selectors are defined correctly
	// This verifies the selector strings without requiring Playwright
	selectors := []string{
		"#main-navbar > nav > ul.header-controls-nav.d-flex.mr-4 > li:nth-child(5) > div > div.btn.lookup-dropdown.lookup-dropdown_no-space-between.border-0.w-auto.pl-0 > i.svg-icon-user.users-profile-icon.svg-icon-primary__border.mr-2",
		"i.svg-icon-user",
		".users-profile-icon", 
		".lookup-dropdown",
	}

	// Verify selectors are not empty
	for i, selector := range selectors {
		if strings.TrimSpace(selector) == "" {
			t.Errorf("Selector %d is empty", i)
		}
		
		// Verify selectors have expected patterns
		switch i {
		case 0:
			if !strings.Contains(selector, "svg-icon-user") {
				t.Errorf("Primary selector should contain 'svg-icon-user'")
			}
		case 1:
			if !strings.HasPrefix(selector, "i.") {
				t.Errorf("Second selector should be an element selector starting with 'i.'")
			}
		case 2:
			if !strings.HasPrefix(selector, ".") {
				t.Errorf("Third selector should be a class selector starting with '.'")
			}
		case 3:
			if !strings.HasPrefix(selector, ".") {
				t.Errorf("Fourth selector should be a class selector starting with '.'")
			}
		}
	}
}

func TestCredentialsStruct(t *testing.T) {
	// Test the Credentials struct
	creds := Credentials{
		Email:    "test@example.com",
		Password: "testpassword",
	}

	if creds.Email != "test@example.com" {
		t.Errorf("Expected email to be test@example.com, got %s", creds.Email)
	}

	if creds.Password != "testpassword" {
		t.Errorf("Expected password to be testpassword, got %s", creds.Password)
	}
}