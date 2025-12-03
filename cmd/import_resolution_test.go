package cmd

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestCheckScreenResolution_BelowMinimum(t *testing.T) {
	// Save original values
	originalWidth := width
	originalHeight := height
	defer func() {
		width = originalWidth
		height = originalHeight
	}()

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// Test with resolution below minimum
	width = 1280
	height = 720
	checkScreenResolution()

	output := buf.String()
	if !strings.Contains(output, "WARNING") {
		t.Errorf("Expected warning for resolution %dx%d, but got no warning. Output: %s", width, height, output)
	}
	if !strings.Contains(output, "1920x1080") {
		t.Errorf("Expected warning to mention recommended resolution 1920x1080. Output: %s", output)
	}
}

func TestCheckScreenResolution_AtMinimum(t *testing.T) {
	// Save original values
	originalWidth := width
	originalHeight := height
	defer func() {
		width = originalWidth
		height = originalHeight
	}()

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// Test with resolution at minimum (should not warn)
	width = 1920
	height = 1080
	checkScreenResolution()

	output := buf.String()
	if strings.Contains(output, "WARNING") {
		t.Errorf("Did not expect warning for resolution %dx%d, but got: %s", width, height, output)
	}
}

func TestCheckScreenResolution_AboveMinimum(t *testing.T) {
	// Save original values
	originalWidth := width
	originalHeight := height
	defer func() {
		width = originalWidth
		height = originalHeight
	}()

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// Test with resolution above minimum (should not warn)
	width = 2560
	height = 1440
	checkScreenResolution()

	output := buf.String()
	if strings.Contains(output, "WARNING") {
		t.Errorf("Did not expect warning for resolution %dx%d, but got: %s", width, height, output)
	}
}
