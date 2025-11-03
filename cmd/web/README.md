# Web Client

This directory contains the web client for the Edubase to PDF server.

## Files

- `index.html` - Single-page application with Tailwind CSS for downloading PDFs

## Features

- Modern, responsive UI built with Tailwind CSS
- Form validation for all fields
- Real-time server health check
- Progress indication during downloads
- Error handling and status messages
- Direct PDF download to browser

## Usage

The web client is automatically served when you start the server:

```bash
edubase-to-pdf server
```

Then open your browser to `http://localhost:8080`

## Architecture

The client is a static HTML file with embedded JavaScript that:
1. Provides a form for entering credentials and book information
2. Makes POST requests to the `/download` endpoint
3. Handles the PDF blob response and triggers a download
4. Includes a health check button to verify server status

## Styling

Uses Tailwind CSS via CDN for:
- Responsive layout
- Form styling
- Button states
- Status messages
- Progress bars

## Security

The Tailwind CSS library is loaded from a versioned CDN with Subresource Integrity (SRI) hash verification. This ensures:
- The loaded script matches the expected content
- Protection against CDN compromise or tampering
- Verification of the script's authenticity

Since this web client is served from a local server (localhost), the primary security concern is the integrity of external resources, which is addressed through SRI.
