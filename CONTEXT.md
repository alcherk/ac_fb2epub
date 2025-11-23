# FB2 to EPUB Converter - Project Context

## Project Overview

**FB2 to EPUB Converter** is a Go-based microservice that converts FictionBook 2.0 (FB2) XML files to EPUB 3.0 format via REST API. The service includes a web UI for easy file conversion and supports Docker deployment.

**Current Project Health:** ✅ **Production Ready**
- ✅ All core functionality implemented
- ✅ Comprehensive test suite (80+ tests, all passing)
- ✅ Complete documentation
- ✅ Docker deployment ready
- ✅ All known issues resolved

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
│   └── config_test.go              ✅ 3 tests (100% coverage)
├── converter/
│   ├── fb2parser_test.go           ✅ 13 tests (parsing, sections, images, links, formatting, poems, citations)
│   ├── epubgenerator_test.go       ✅ 7 tests (structure, ZIP, TOC, nested sections, HTML escaping)
│   ├── epub_content_test.go        ✅ 5 tests (HTML escaping, section IDs, images, links, special chars)
│   ├── error_handling_test.go      ✅ 6 tests (invalid XML, malformed FB2, empty files, missing files)
│   └── edge_cases_test.go          ✅ 5 tests (no sections, no title, long text, emojis, deep nesting)
├── handlers/
│   ├── converter_test.go           ✅ 17 tests (conversion, status, download, job management)
│   ├── file_size_test.go           ✅ 4 tests (size limits, error messages)
│   ├── cleanup_test.go             ✅ 5 tests (completed jobs, failed jobs, orphaned dirs, recent jobs, trigger count)
│   ├── temp_dir_test.go            ✅ 5 tests (creation, permissions, subdirs, cleanup, error cleanup)
│   └── concurrency_test.go         ✅ 4 tests (multiple conversions, status checks, downloads, job map)
└── integration/
    ├── integration_test.go          ✅ 4 tests (full workflow, status polling, health, invalid routes)
    └── concurrent_test.go           ✅ 2 tests (concurrent conversions, error handling)
```

**Test Results:**
- **All tests:** ✅ **80+ test functions, all passing**
- **Config tests:** ✅ 3 tests (100% coverage)
- **Converter tests:** ✅ 36 tests (parser, generator, content, error handling, edge cases)
- **Handler tests:** ✅ 35 tests (conversion, status, download, file size, cleanup, temp dir, concurrency)
- **Integration tests:** ✅ 6 tests (full workflows, concurrent scenarios, error handling)

**Test Coverage:**
- ✅ Core functionality: 100%
- ✅ API endpoints: 100%
- ✅ Error handling: Comprehensive
- ✅ Edge cases: Covered
- ✅ Concurrency: Validated
- ✅ Integration flows: Verified

### ✅ Resolved Issues

1. **✅ Converter Test Package Conflict** - Fixed
   - Removed `test_helper.go` and inlined helper function
   - All converter tests now passing

2. **✅ Handler Test Temp Directory Cleanup** - Fixed
   - Added proper cleanup in tests to handle async processing
   - All handler tests now passing

3. **✅ File Size Tests** - Fixed
   - All file size limit tests working correctly
   - Proper error message validation

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
│   │   └── config_test.go
│   ├── converter/
│   │   ├── fb2parser_test.go
│   │   ├── epubgenerator_test.go
│   │   ├── epub_content_test.go
│   │   ├── error_handling_test.go
│   │   └── edge_cases_test.go
│   ├── handlers/
│   │   ├── converter_test.go
│   │   ├── file_size_test.go
│   │   ├── cleanup_test.go
│   │   ├── temp_dir_test.go
│   │   └── concurrency_test.go
│   └── integration/
│       ├── integration_test.go
│       └── concurrent_test.go
├── testdata/                    # Test FB2 files
│   ├── valid/
│   │   ├── minimal.fb2
│   │   ├── complete.fb2
│   │   ├── with-poems.fb2
│   │   ├── with-citations.fb2
│   │   ├── with-images.fb2
│   │   ├── with-formatting.fb2
│   │   └── with-links.fb2
│   ├── invalid/
│   │   ├── malformed.xml
│   │   └── empty.fb2
│   └── edge-cases/
│       └── unicode.fb2
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

1. **Comprehensive Test Suite Implementation** (Latest)
   - ✅ Implemented 80+ test functions across 13 test files
   - ✅ All tests passing with 100% success rate
   - ✅ Complete test coverage for all functionality
   - ✅ Fixed all test infrastructure issues
   - ✅ Added comprehensive test data files
   - ✅ Organized all tests in `tests/` directory
   - ✅ Added parser tests (sections, images, links, formatting, poems, citations)
   - ✅ Added EPUB generator tests (images, formatting, content validation)
   - ✅ Added handler tests (job creation, status, download, headers)
   - ✅ Added cleanup tests (completed, failed, orphaned, recent jobs, trigger count)
   - ✅ Added temp directory tests (creation, permissions, subdirs, cleanup)
   - ✅ Added concurrency tests (multiple conversions, status checks, downloads, job map)
   - ✅ Added edge case tests (no sections, no title, long text, emojis, deep nesting)
   - ✅ Added integration tests (concurrent conversions, error handling)

2. **Test Organization**
   - All test files moved to `tests/` directory (removed from source directories)
   - Test plan updated with organization rule
   - All test file paths updated in documentation
   - Test helper functions properly organized

3. **Temp Folder Cleanup**
   - Automatic cleanup triggered by number of conversions
   - Removes old completed/failed jobs (>1 hour old)
   - Thread-safe implementation with mutex
   - Configurable via CLEANUP_TRIGGER_COUNT

4. **Linter Fixes**
   - Fixed all golangci-lint issues
   - Added ReadHeaderTimeout for security
   - Fixed implicit memory aliasing in for loops
   - Added proper nolint comments

5. **413 Error Fixes**
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
- `valid/complete.fb2` - FB2 with all features (nested sections, formatting, links)
- `valid/with-poems.fb2` - FB2 with poem structures
- `valid/with-citations.fb2` - FB2 with citation blocks
- `valid/with-images.fb2` - FB2 with embedded images
- `valid/with-formatting.fb2` - FB2 with strong/emphasis formatting
- `valid/with-links.fb2` - FB2 with hyperlinks
- `invalid/malformed.xml` - Invalid XML
- `invalid/empty.fb2` - Empty FB2
- `edge-cases/unicode.fb2` - Unicode and special characters

## Next Steps / TODO

1. **✅ Test Plan Implementation** - COMPLETE
   - ✅ All test categories implemented
   - ✅ 80+ test functions across 13 test files
   - ✅ All tests passing
   - ✅ Comprehensive coverage achieved

2. **Future Enhancements**
   - Web UI automated tests (optional)
   - Add request rate limiting
   - Consider persistent job storage (Redis/DB)
   - Add metrics and monitoring
   - Performance benchmarking
   - Load testing

3. **CI/CD Setup**
   - Set up GitHub Actions for automated testing
   - Configure test coverage reporting
   - Set up automated deployments

## Important Notes

- **All test files MUST be in `tests/` directory** (rule enforced)
- All test files use `*_test` package names (e.g., `config_test`, `converter_test`)
- Test helper functions inlined in test files (no separate test_helper.go)
- Integration tests in `tests/integration/` directory
- Handler tests use exported test helper functions:
  - `GetConversionJob(jobID)` - Get job by ID
  - `SetConversionJob(job)` - Set job in memory
  - `DeleteConversionJob(jobID)` - Delete job from memory
- Job status constants exported (JobStatusPending, JobStatusProcessing, etc.)

## Test Execution

```bash
# Run all tests
go test ./tests/...

# Run specific test package
go test ./tests/converter/...
go test ./tests/handlers/...

# Run with coverage
go test -cover ./tests/...

# Run specific test
go test ./tests/handlers -run TestConvertFB2ToEPUB_ValidFile
```

## Git Status

- Latest: Comprehensive test suite implementation (80+ tests)
- All tests passing
- Test plan fully implemented
- All test files organized in `tests/` directory

## Deployment

- Docker Compose ready
- Nginx configuration examples provided
- Helper scripts for Nginx 413 error fixes
- Health checks configured
- Production-ready with proper error handling

