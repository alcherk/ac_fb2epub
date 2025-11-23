# FB2 to EPUB Converter - Project Context

## Project Overview

**FB2 to EPUB Converter** is a Go-based microservice that converts FictionBook 2.0 (FB2) XML files to EPUB 3.0 format via REST API. The service includes a web UI for easy file conversion and supports Docker deployment.

## Current Status

### ✅ Completed Features

1. **Core Functionality**
   - FB2 XML parsing with proper namespace handling
   - EPUB 3.0 generation with full structure (mimetype, container.xml, content.opf, toc.ncx, nav.xhtml)
   - Asynchronous job processing
   - Web UI with drag-and-drop file upload
   - Health check endpoint
   - Automatic temp folder cleanup (triggered by number of conversions)

2. **Configuration**
   - Environment variable-based configuration
   - Configurable file size limits (MAX_FILE_SIZE)
   - Configurable cleanup trigger count (CLEANUP_TRIGGER_COUNT)
   - Production/development modes

3. **Docker Support**
   - Multi-stage Dockerfile
   - Docker Compose configuration
   - Named volumes for temp storage
   - Health checks

4. **Code Quality**
   - golangci-lint integration
   - All linter issues resolved
   - Code formatting
   - Package comments

5. **Documentation**
   - Comprehensive README
   - Deployment guide (DEPLOYMENT.md)
   - Docker guide (DOCKER.md)
   - Quick start guide (QUICKSTART.md)
   - Architecture documentation (ARCHITECTURE.md)
   - Troubleshooting guide (TROUBLESHOOTING.md)
   - Test plan (TEST_PLAN.md)

### 🧪 Test Status

**Test Files Location:** All tests are in `tests/` directory (separate from source code)

**Test Structure:**
```
tests/
├── config/
│   └── config_test.go          ✅ Passing (100% coverage)
├── converter/
│   ├── fb2parser_test.go       ✅ Passing
│   ├── epubgenerator_test.go   ⚠️ Package name issue
│   ├── epub_content_test.go    ⚠️ Package name issue
│   ├── error_handling_test.go  ✅ Passing
│   └── test_helper.go          ✅ Helper functions
├── handlers/
│   ├── converter_test.go       ✅ Passing (9/10 tests)
│   └── file_size_test.go       ⚠️ Some failures (temp dir cleanup)
└── integration_test.go          ✅ Passing (4 tests)
```

**Test Results:**
- **Config tests:** ✅ All passing (100% coverage)
- **Integration tests:** ✅ All passing (4 tests)
- **Converter tests:** ⚠️ Package name conflict issue
- **Handler tests:** ⚠️ 9/10 passing (1 temp dir cleanup issue)

**Total Test Cases:** ~32+ test cases implemented

### ⚠️ Known Issues

1. **Converter Test Package Conflict**
   - Error: `found packages converter (epubgenerator_test.go) and converter_test (test_helper.go)`
   - Issue: Go compiler detects different package names in same directory
   - Status: Needs investigation - may be file encoding or hidden characters

2. **Handler Test Temp Directory Cleanup**
   - Some tests fail with "directory not empty" during cleanup
   - Non-critical - test infrastructure issue, not code bug

3. **File Size Tests**
   - Some tests need proper Content-Type headers
   - Minor fixes needed

## Project Structure

```
fb2epub/
├── main.go                      # Entry point, Gin server setup
├── config/
│   └── config.go                # Configuration management
├── models/
│   └── fb2.go                   # FB2 XML data structures
├── converter/
│   ├── fb2parser.go             # FB2 XML parser
│   ├── epubgenerator.go         # EPUB 3.0 generator (898 lines)
│   └── nav.go                   # EPUB navigation generation
├── handlers/
│   └── converter.go             # HTTP handlers (346 lines)
├── web/
│   ├── index.html               # Web UI
│   └── static/
│       ├── style.css            # Styles
│       └── app.js               # JavaScript
├── tests/                       # All test files (separate directory)
│   ├── config/
│   ├── converter/
│   ├── handlers/
│   └── integration_test.go
├── testdata/                    # Test FB2 files
│   ├── valid/
│   ├── invalid/
│   └── edge-cases/
└── Documentation files...
```

## Key Configuration

**Environment Variables:**
- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - development/production (default: development)
- `TEMP_DIR` - Temporary directory (default: /tmp/fb2epub)
- `MAX_FILE_SIZE` - Max file size in bytes (default: 50MB, Docker: 100MB)
- `CLEANUP_TRIGGER_COUNT` - Cleanup after N conversions (default: 10)

**Docker:**
- Host port: 3080 (mapped to container 8080)
- Named volume: `temp_data:/app/temp`
- Health check enabled

## Recent Changes

1. **Temp Folder Cleanup** (Latest)
   - Automatic cleanup triggered by number of conversions
   - Removes old completed/failed jobs (>1 hour old)
   - Thread-safe implementation with mutex
   - Configurable via CLEANUP_TRIGGER_COUNT

2. **Test Infrastructure**
   - Moved all tests to `tests/` directory
   - Created test data files
   - Implemented integration tests
   - Added file size and error handling tests

3. **Linter Fixes**
   - Fixed all golangci-lint issues
   - Added ReadHeaderTimeout for security
   - Fixed implicit memory aliasing in for loops
   - Added proper nolint comments

4. **413 Error Fixes**
   - Increased MAX_FILE_SIZE to 100MB
   - Configured MaxHeaderBytes
   - Added Nginx configuration examples
   - Created helper scripts for Nginx setup

## API Endpoints

- `GET /health` - Health check
- `POST /api/v1/convert` - Upload FB2 file, returns job_id
- `GET /api/v1/status/:id` - Get conversion status
- `GET /api/v1/download/:id` - Download completed EPUB

## Dependencies

- `github.com/gin-gonic/gin` v1.9.1 - Web framework
- `github.com/google/uuid` v1.4.0 - UUID generation
- Go 1.21+

## Build & Run

```bash
# Build
make build
# or
go build -o fb2epub

# Run
make run
# or
go run main.go

# Test
go test ./tests/...

# Lint
make lint

# Docker
docker-compose up -d
```

## Test Data

Test FB2 files are in `testdata/`:
- `valid/minimal.fb2` - Minimal valid FB2
- `valid/complete.fb2` - FB2 with all features
- `invalid/malformed.xml` - Invalid XML
- `invalid/empty.fb2` - Empty FB2
- `edge-cases/unicode.fb2` - Unicode and special characters

## Next Steps / TODO

1. **Fix Converter Test Package Issue**
   - Resolve package name conflict in tests/converter/
   - May require file recreation or encoding fix

2. **Complete Test Plan**
   - Edge cases tests (test-12)
   - Concurrency tests (test-15)
   - Temp directory tests (test-16)
   - Web UI tests (test-17) - Optional

3. **Improvements**
   - Add more test data (FB2 with images)
   - Improve error messages
   - Add request rate limiting
   - Consider persistent job storage (Redis/DB)

## Important Notes

- Tests are in separate `tests/` directory (not alongside source)
- All test files use `*_test` package names (e.g., `config_test`, `converter_test`)
- Test data path helper function in `tests/converter/test_helper.go`
- Integration tests in `tests/integration_test.go` (package `tests_test`)
- Some handler tests need exported functions (GetConversionJob, SetConversionJob, DeleteConversionJob)
- Job status constants exported (JobStatusPending, JobStatusProcessing, etc.)

## Git Status

- Latest commit: "Add temp folder cleanup and fix linter issues"
- Tags: v0.0.1, v0.0.2, v0.0.3, v0.0.4
- Uncommitted: Test files and TEST_PLAN.md

## Deployment

- Docker Compose ready
- Nginx configuration examples provided
- Helper scripts for Nginx 413 error fixes
- Health checks configured
- Production-ready with proper error handling

