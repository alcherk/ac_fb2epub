# Quick Start Guide

## Local Development (macOS)

### 1. Install Dependencies

```bash
go mod download
```

### 2. Build

```bash
make build
# or
go build -o fb2epub
```

### 3. Run

**Recommended for macOS:**
```bash
go run main.go
```

Or use the binary:
```bash
./fb2epub
```

**Note:** If you encounter `dyld: missing LC_UUID` errors on macOS, use `go run main.go` instead.

The service will start on `http://localhost:8080`

### 4. Access Web UI

Open your browser and navigate to:
```
http://localhost:8080
```

The web UI provides a beautiful interface for:
- Drag and drop file upload
- Real-time conversion status
- Direct EPUB download

### 5. Test

In another terminal:

```bash
# Health check
curl http://localhost:8080/health

# Convert a file (if you have one)
curl -X POST http://localhost:8080/api/v1/convert -F "file=@your-book.fb2"
```

## Using the Test Script

```bash
# Make sure service is running first
./fb2epub &

# Test with a file
./test.sh path/to/your-book.fb2
```

## Environment Variables

You can customize the service behavior:

```bash
export PORT=8080                    # Server port
export ENVIRONMENT=development      # development or production
export TEMP_DIR=/tmp/fb2epub        # Temporary directory
export MAX_FILE_SIZE=52428800       # Max file size in bytes (50MB)
export CLEANUP_TRIGGER_COUNT=10     # Cleanup temp folder after N conversions

./fb2epub
```

## Example Workflow

1. **Start the service:**
   ```bash
   ./fb2epub
   ```

2. **In another terminal, convert a file:**
   ```bash
   # Upload and get job ID
   curl -X POST http://localhost:8080/api/v1/convert \
     -F "file=@book.fb2" > response.json
   
   # Extract job ID (using jq)
   JOB_ID=$(cat response.json | jq -r '.job_id')
   
   # Check status
   curl http://localhost:8080/api/v1/status/$JOB_ID
   
   # Download when ready
   curl -O http://localhost:8080/api/v1/download/$JOB_ID
   ```

## Troubleshooting

### Port already in use
```bash
export PORT=8081
./fb2epub
```

### Permission denied
```bash
chmod +x fb2epub
mkdir -p /tmp/fb2epub
```

### Service not responding
Check if it's running:
```bash
curl http://localhost:8080/health
```

## Docker Quick Start

### Using Docker Compose

```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

### Using Docker

```bash
# Build image
docker build -t fb2epub .

# Run container
docker run -d --name fb2epub -p 3080:8080 fb2epub

# View logs
docker logs -f fb2epub
```

Access the web UI at `http://localhost:3080`

## Next Steps

- See [README.md](README.md) for detailed documentation
- See [DEPLOYMENT.md](DEPLOYMENT.md) for VPS deployment instructions
- See [DOCKER.md](DOCKER.md) for Docker deployment guide

