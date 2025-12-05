# HTTP Server API Documentation

## Overview

The HTTP server provides a stateless REST API for downloading PDFs from Edubase. The server handles authentication, page downloading, and PDF generation, streaming the result directly to clients.

## Web Client

The server includes a built-in web client for easy interaction with the API. Simply start the server and navigate to `http://localhost:8080` in your browser to access the web interface.

**Features:**
- Modern, responsive UI built with Tailwind CSS
- Form-based PDF download with validation
- Real-time server health check
- Progress indication during download
- Error handling and status messages

## Starting the Server

```bash
# Start with default settings (host: 0.0.0.0, port: 8080)
edubase-to-pdf server

# Start with custom port
edubase-to-pdf server --port 9090

# Start with custom host and port
edubase-to-pdf server --host localhost --port 3000
```

Once started, access the web client at: `http://localhost:8080`

## API Endpoints

### Web Client

Access the browser-based client interface.

**Endpoint:** `GET /`

**Response:** HTML page with interactive form

### Health Check

Check if the server is running and healthy.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "ok"
}
```

**Status Codes:**
- `200 OK`: Server is healthy

### Download PDF

Download a book from Edubase as a PDF file.

**Endpoint:** `POST /download`

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "email": "your_email@example.com",
  "password": "your_password",
  "book_id": 12345,
  "start_page": 1,
  "max_pages": -1
}
```

**Request Parameters:**
- `email` (string, required): Your Edubase account email
- `password` (string, required): Your Edubase account password
- `book_id` (integer, required): The ID of the book to download (must be > 0)
- `start_page` (integer, required): The page to start downloading from (must be > 0)
- `max_pages` (integer, required): Maximum number of pages to download
  - `-1`: Download all pages
  - `> 0`: Download specified number of pages
  - Cannot be `0`

**Response (Success):**
- **Status Code:** `200 OK`
- **Headers:**
  - `Content-Type: application/pdf`
  - `Content-Disposition: attachment; filename=book_{book_id}.pdf`
  - `Content-Length: {size_in_bytes}`
- **Body:** Binary PDF file stream

**Error Responses:**

| Status Code | Error Code | Description |
|------------|------------|-------------|
| 400 Bad Request | `invalid_json` | Request body is not valid JSON |
| 400 Bad Request | `validation_error` | One or more request parameters are invalid |
| 401 Unauthorized | `auth_failed` | Authentication with Edubase failed |
| 405 Method Not Allowed | `method_not_allowed` | Wrong HTTP method used |
| 500 Internal Server Error | `processing_error` | Server encountered an error during processing |

**Error Response Format:**
```json
{
  "error": "error_code",
  "message": "Human-readable error message"
}
```

## Examples

### Using cURL

#### Health Check
```bash
curl http://localhost:8080/health
```

#### Download a PDF
```bash
curl -X POST http://localhost:8080/download \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "mypassword",
    "book_id": 12345,
    "start_page": 1,
    "max_pages": -1
  }' \
  --output book.pdf
```

#### Download specific pages
```bash
curl -X POST http://localhost:8080/download \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "mypassword",
    "book_id": 12345,
    "start_page": 5,
    "max_pages": 10
  }' \
  --output book_pages_5-14.pdf
```

### Using Python

```python
import requests

# Download PDF
response = requests.post('http://localhost:8080/download', json={
    'email': 'user@example.com',
    'password': 'mypassword',
    'book_id': 12345,
    'start_page': 1,
    'max_pages': -1
})

if response.status_code == 200:
    with open('book.pdf', 'wb') as f:
        f.write(response.content)
    print('PDF downloaded successfully')
else:
    print(f'Error: {response.json()}')
```

### Using JavaScript (Node.js)

```javascript
const fetch = require('node-fetch');
const fs = require('fs');

async function downloadPDF() {
  const response = await fetch('http://localhost:8080/download', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      email: 'user@example.com',
      password: 'mypassword',
      book_id: 12345,
      start_page: 1,
      max_pages: -1
    })
  });

  if (response.ok) {
    const buffer = await response.buffer();
    fs.writeFileSync('book.pdf', buffer);
    console.log('PDF downloaded successfully');
  } else {
    const error = await response.json();
    console.error('Error:', error);
  }
}

downloadPDF();
```

## Validation Rules

The server performs strict validation on all request parameters:

1. **Email**
   - Must not be empty
   - Any valid string accepted (server doesn't validate email format)

2. **Password**
   - Must not be empty
   - Any valid string accepted

3. **Book ID**
   - Must be a positive integer (> 0)
   - Cannot be 0 or negative

4. **Start Page**
   - Must be a positive integer (> 0)
   - Cannot be 0 or negative

5. **Max Pages**
   - Must be -1 (all pages) or a positive integer (> 0)
   - Cannot be 0

## Server Configuration

The server is configured with reasonable defaults:

- **Read Timeout:** 15 seconds
- **Write Timeout:** 5 minutes (allows for large PDF generation)
- **Idle Timeout:** 60 seconds
- **Default Host:** 0.0.0.0 (listens on all interfaces)
- **Default Port:** 8080

## Security Considerations

1. **Credentials**: User credentials are sent in the request body. Use HTTPS in production to encrypt traffic.
2. **Stateless**: Each request is independent - no session management or credential caching.
3. **Temporary Files**: Screenshots are stored in temporary directories and cleaned up after each request.
4. **Resource Management**: Browser instances are properly cleaned up after each request.

## Error Handling

The server provides detailed error messages for common issues:

- **Invalid JSON**: Check that your request body is properly formatted JSON
- **Missing Fields**: Ensure all required fields are present
- **Invalid Values**: Verify that numeric fields contain valid values
- **Authentication Failures**: Check your Edubase credentials
- **Processing Errors**: May occur due to network issues, invalid book IDs, or Edubase being unavailable

## Rate Limiting

The server does not implement rate limiting. Consider adding a reverse proxy (like Nginx) with rate limiting in production environments.

## Graceful Shutdown

The server handles SIGINT and SIGTERM signals gracefully, allowing ongoing requests to complete before shutting down (with a 30-second timeout).

```bash
# Press Ctrl+C to stop the server
# Or send SIGTERM
kill -TERM <pid>
```
