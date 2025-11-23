# FB2 to EPUB Converter Service

A Go-based microservice that converts FictionBook 2.0 (FB2) files to EPUB format via REST API.

## Features

- **Web UI** - Beautiful, modern web interface for easy file conversion
- RESTful API for FB2 to EPUB conversion
- Asynchronous job processing
- Health check endpoint
- Configurable via environment variables
- Cross-platform support (macOS, Linux)
- **Docker support** - Easy deployment with Docker and Docker Compose

## Architecture

```
fb2epub/
├── main.go              # Application entry point
├── config/              # Configuration management
├── models/              # Data models (FB2 structure)
├── converter/           # Core conversion logic
│   ├── fb2parser.go     # FB2 XML parser
│   └── epubgenerator.go # EPUB file generator
├── handlers/            # HTTP request handlers
└── web/                 # Web UI
    ├── index.html       # Main HTML page
    └── static/          # Static assets
        ├── style.css    # Styles
        └── app.js       # JavaScript
```

## Local Development (macOS)

### Prerequisites

- Go 1.21 or later
- Make (optional, for convenience commands)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd fb2epub
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o fb2epub
```

Or use Make:
```bash
make build
```

### Running Locally

1. Set environment variables (optional):
```bash
export PORT=8080
export ENVIRONMENT=development
export TEMP_DIR=/tmp/fb2epub
export MAX_FILE_SIZE=52428800  # 50MB in bytes
```

2. Run the service:

**Recommended: Use Docker (avoids macOS issues):**
```bash
docker-compose up -d
```

**Or use the run script:**
```bash
./run.sh
```

**Or manually:**
```bash
go run main.go
```

**Note:** If you encounter `dyld: missing LC_UUID` errors on macOS, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for solutions. Docker is the most reliable option.

The service will start on `http://localhost:8080`

### Using the Web UI

Once the service is running, open your browser and navigate to:
```
http://localhost:8080
```

The web UI provides:
- Drag and drop file upload
- Real-time conversion status
- Progress indicator
- Direct EPUB download
- Beautiful, modern interface

Simply drag your FB2 file onto the upload area or click to browse, and the conversion will start automatically.

### Docker Deployment

#### Using Docker Compose (Recommended)

1. Build and run with Docker Compose:
```bash
docker-compose up -d
```

2. Access the web UI at `http://localhost:3080`

3. View logs:
```bash
docker-compose logs -f
```

4. Stop the service:
```bash
docker-compose down
```

#### Using Docker

1. Build the Docker image:
```bash
docker build -t fb2epub .
```

2. Run the container:
```bash
docker run -d \
  --name fb2epub \
  -p 3080:8080 \
  -v $(pwd)/temp:/app/temp \
  fb2epub
```

3. Access the web UI at `http://localhost:3080`

4. View logs:
```bash
docker logs -f fb2epub
```

5. Stop the container:
```bash
docker stop fb2epub
docker rm fb2epub
```

### Testing the Service

#### 1. Health Check
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "service": "fb2epub"
}
```

#### 2. Convert FB2 to EPUB

Upload an FB2 file:
```bash
curl -X POST http://localhost:8080/api/v1/convert \
  -F "file=@your-book.fb2"
```

Response:
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "message": "Conversion started"
}
```

#### 3. Check Conversion Status

```bash
curl http://localhost:8080/api/v1/status/{job_id}
```

Response (when completed):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "created_at": "2024-01-15T10:30:00Z",
  "download_url": "/api/v1/download/550e8400-e29b-41d4-a716-446655440000"
}
```

#### 4. Download EPUB

```bash
curl -O http://localhost:8080/api/v1/download/{job_id}
```

Or open in browser:
```
http://localhost:8080/api/v1/download/{job_id}
```

### Example with Sample FB2 File

If you have a sample FB2 file:

```bash
# Start conversion
JOB_ID=$(curl -s -X POST http://localhost:8080/api/v1/convert \
  -F "file=@sample.fb2" | jq -r '.job_id')

# Wait a moment for processing
sleep 2

# Check status
curl http://localhost:8080/api/v1/status/$JOB_ID

# Download when ready
curl -O http://localhost:8080/api/v1/download/$JOB_ID
```

## API Endpoints

### POST /api/v1/convert
Upload an FB2 file for conversion.

**Request:**
- Content-Type: `multipart/form-data`
- Field name: `file`
- File extension: `.fb2` or `.xml`

**Response:**
```json
{
  "job_id": "uuid",
  "status": "processing",
  "message": "Conversion started"
}
```

### GET /api/v1/status/:id
Get the status of a conversion job.

**Response (processing):**
```json
{
  "id": "uuid",
  "status": "processing",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Response (completed):**
```json
{
  "id": "uuid",
  "status": "completed",
  "created_at": "2024-01-15T10:30:00Z",
  "download_url": "/api/v1/download/uuid"
}
```

**Response (failed):**
```json
{
  "id": "uuid",
  "status": "failed",
  "created_at": "2024-01-15T10:30:00Z",
  "error": "Error message"
}
```

### GET /api/v1/download/:id
Download the converted EPUB file.

**Response:**
- Content-Type: `application/epub+zip`
- File download

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "service": "fb2epub"
}
```

## Configuration

Environment variables:

- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - Environment mode: development/production (default: development)
- `TEMP_DIR` - Temporary directory for file processing (default: /tmp/fb2epub)
- `MAX_FILE_SIZE` - Maximum file size in bytes (default: 52428800 = 50MB)

## Project Structure

```
.
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── config/
│   └── config.go          # Configuration loader
├── models/
│   └── fb2.go             # FB2 data structures
├── converter/
│   ├── fb2parser.go       # FB2 XML parser
│   └── epubgenerator.go   # EPUB generator
├── handlers/
│   └── converter.go       # HTTP handlers
├── Makefile               # Build automation
├── .gitignore             # Git ignore rules
└── README.md              # This file
```

## Building

```bash
# Build for current platform
go build -o fb2epub

# Build for Linux (for VPS deployment)
GOOS=linux GOARCH=amd64 go build -o fb2epub-linux

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o fb2epub-macos
```

## Code Quality

### Linting

The project uses [golangci-lint](https://golangci-lint.run/) for code quality checks.

```bash
# Run linter
make lint

# Run linter with auto-fix
make lint-fix

# Check code formatting
make fmt-check

# Format code
make fmt
```

### Installing golangci-lint

If you don't have golangci-lint installed, the Makefile will automatically install the latest version. You can also install it manually:

```bash
# Using the install script (recommended)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest

# Or using Homebrew (macOS)
brew install golangci-lint

# Or using go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

The linter configuration is in `.golangci.yml`. The project uses golangci-lint v2.x which requires Go 1.21+.

### Pre-commit Checks

Before committing, it's recommended to run:

```bash
make fmt      # Format code
make lint     # Check for issues
make test     # Run tests
```

## Troubleshooting

### Port already in use
Change the port:
```bash
export PORT=8081
./fb2epub
```

### Permission denied
Make sure the temp directory is writable:
```bash
mkdir -p /tmp/fb2epub
chmod 755 /tmp/fb2epub
```

### File too large
Increase MAX_FILE_SIZE:
```bash
export MAX_FILE_SIZE=104857600  # 100MB
```

## License

MIT

