# Test Plan for FB2 to EPUB Converter

## Overview

This document outlines a comprehensive test plan for the FB2 to EPUB converter service. The plan covers unit tests, integration tests, API endpoint tests, and edge case testing.

## Test Structure

```
fb2epub/
├── testdata/              # Test data files
│   ├── valid/
│   │   ├── minimal.fb2    # Minimal valid FB2 file
│   │   ├── complete.fb2   # FB2 with all features
│   │   ├── with-images.fb2
│   │   └── nested-sections.fb2
│   ├── invalid/
│   │   ├── malformed.xml
│   │   ├── empty.fb2
│   │   └── missing-elements.fb2
│   └── edge-cases/
│       ├── unicode.fb2
│       ├── large.fb2
│       └── deep-nesting.fb2
├── config/
│   └── config_test.go
├── converter/
│   ├── fb2parser_test.go
│   └── epubgenerator_test.go
├── handlers/
│   └── converter_test.go
└── main_test.go           # Integration tests
```

## Test Categories

### 1. Unit Tests

#### 1.1 Config Package Tests (`config/config_test.go`)

**Test Cases:**
- `TestLoad_DefaultValues`: Verify default configuration values
- `TestLoad_EnvironmentVariables`: Test loading from environment variables
  - PORT
  - ENVIRONMENT
  - TEMP_DIR
  - MAX_FILE_SIZE
  - CLEANUP_TRIGGER_COUNT
- `TestLoad_InvalidValues`: Test handling of invalid environment values
  - Negative MAX_FILE_SIZE
  - Negative CLEANUP_TRIGGER_COUNT
  - Empty strings
  - Non-numeric values

**Test Data:** None required (uses environment variables)

#### 1.2 FB2 Parser Tests (`converter/fb2parser_test.go`)

**Test Cases:**
- `TestParseFB2_ValidFile`: Parse valid FB2 file
- `TestParseFB2_InvalidXML`: Handle malformed XML
- `TestParseFB2_MissingFile`: Handle file not found
- `TestParseFB2_EmptyFile`: Handle empty file
- `TestParseFB2_WithSections`: Parse FB2 with nested sections
- `TestParseFB2_WithImages`: Parse FB2 with embedded images
- `TestParseFB2_WithLinks`: Parse FB2 with links
- `TestParseFB2_WithFormatting`: Parse FB2 with strong/emphasis
- `TestParseFB2_WithPoems`: Parse FB2 with poems
- `TestParseFB2_WithCitations`: Parse FB2 with citations
- `TestParseFB2FromReader`: Test reader-based parsing

**Test Data:** 
- `testdata/valid/minimal.fb2`
- `testdata/valid/complete.fb2`
- `testdata/invalid/malformed.xml`
- `testdata/invalid/empty.fb2`

#### 1.3 EPUB Generator Tests (`converter/epubgenerator_test.go`)

**Test Cases:**
- `TestGenerateEPUB_ValidStructure`: Verify EPUB structure
  - mimetype file exists and is first
  - META-INF/container.xml exists
  - OEBPS/content.opf exists
  - OEBPS/toc.ncx exists
  - OEBPS/nav.xhtml exists
- `TestGenerateEPUB_ValidZIP`: Verify EPUB is valid ZIP archive
- `TestGenerateEPUB_WithImages`: Test image embedding
- `TestGenerateEPUB_WithTOC`: Test table of contents generation
- `TestGenerateEPUB_WithNestedSections`: Test nested section handling
- `TestGenerateEPUB_WithFormatting`: Test inline formatting preservation
- `TestGenerateEPUB_HTMLEscaping`: Test HTML escaping in content
- `TestGenerateEPUB_SectionIDs`: Test section ID generation and escaping

**Test Data:** Use parsed FB2 structs from parser tests

### 2. Handler Tests

#### 2.1 Converter Handler Tests (`handlers/converter_test.go`)

**Test Cases:**

**Convert Endpoint:**
- `TestConvertFB2ToEPUB_ValidFile`: Upload valid FB2 file
- `TestConvertFB2ToEPUB_InvalidFileType`: Reject non-FB2 files
- `TestConvertFB2ToEPUB_FileTooLarge`: Reject files exceeding MAX_FILE_SIZE
- `TestConvertFB2ToEPUB_MissingFile`: Handle missing file in request
- `TestConvertFB2ToEPUB_JobCreation`: Verify job is created with correct status
- `TestConvertFB2ToEPUB_JobID`: Verify job ID is returned

**Status Endpoint:**
- `TestGetConversionStatus_ValidJob`: Get status of valid job
- `TestGetConversionStatus_NonExistent`: Handle non-existent job
- `TestGetConversionStatus_Completed`: Verify download_url in response
- `TestGetConversionStatus_Failed`: Verify error message in response
- `TestGetConversionStatus_Processing`: Verify processing status

**Download Endpoint:**
- `TestDownloadEPUB_CompletedJob`: Download completed EPUB
- `TestDownloadEPUB_NonExistentJob`: Handle non-existent job
- `TestDownloadEPUB_NotCompleted`: Reject download for incomplete job
- `TestDownloadEPUB_Headers`: Verify correct HTTP headers
- `TestDownloadEPUB_ValidFile`: Verify downloaded file is valid EPUB

**Cleanup Tests:**
- `TestCleanupOldJobs_CompletedJobs`: Remove old completed jobs
- `TestCleanupOldJobs_FailedJobs`: Remove old failed jobs
- `TestCleanupOldJobs_OrphanedDirs`: Remove directories not in memory
- `TestCleanupOldJobs_RecentJobs`: Don't remove recent jobs
- `TestCleanupOldJobs_TriggerCount`: Test cleanup trigger mechanism
- `TestCleanupOldJobs_Concurrency`: Test thread safety

**Test Data:** 
- Mock HTTP requests using `httptest` package
- Test FB2 files from `testdata/valid/`
- Mock file system for temp directory tests

### 3. Integration Tests

#### 3.1 Full Conversion Flow (`main_test.go`)

**Test Cases:**
- `TestIntegration_FullConversion`: Complete workflow test
  1. Start test server
  2. Upload FB2 file
  3. Poll status until completed
  4. Download EPUB
  5. Verify EPUB is valid
- `TestIntegration_ConcurrentConversions`: Multiple simultaneous conversions
- `TestIntegration_StatusPolling`: Test status polling mechanism
- `TestIntegration_ErrorHandling`: Test error scenarios end-to-end

**Test Data:** 
- `testdata/valid/complete.fb2`
- Test server using `httptest.NewServer()`

### 4. API Endpoint Tests

#### 4.1 HTTP Endpoint Tests

**Test Cases:**
- `TestHealthEndpoint`: Test `/health` endpoint
  - Returns 200 OK
  - Returns correct JSON structure
- `TestConvertEndpoint`: Test `/api/v1/convert`
  - Accepts POST requests
  - Returns 202 Accepted
  - Returns job_id
- `TestStatusEndpoint`: Test `/api/v1/status/:id`
  - Accepts GET requests
  - Returns correct status
- `TestDownloadEndpoint`: Test `/api/v1/download/:id`
  - Accepts GET requests
  - Returns EPUB file
  - Sets correct headers
- `TestInvalidRoutes`: Test 404 for invalid routes

**Test Data:** Mock HTTP requests

### 5. Edge Cases and Error Handling

#### 5.1 Edge Cases

**Test Cases:**
- `TestEdgeCase_EmptyFB2`: Handle empty FB2 file
- `TestEdgeCase_NoSections`: Handle FB2 with no sections
- `TestEdgeCase_NoTitle`: Handle FB2 with no title
- `TestEdgeCase_LongText`: Handle very long text content
- `TestEdgeCase_Unicode`: Handle Unicode characters
- `TestEdgeCase_Emojis`: Handle emoji characters
- `TestEdgeCase_DeepNesting`: Handle deeply nested sections
- `TestEdgeCase_ManyImages`: Handle many embedded images
- `TestEdgeCase_SpecialCharacters`: Handle special XML characters

**Test Data:**
- `testdata/edge-cases/unicode.fb2`
- `testdata/edge-cases/large.fb2`
- `testdata/edge-cases/deep-nesting.fb2`

#### 5.2 Error Handling

**Test Cases:**
- `TestError_InvalidXML`: Handle invalid XML structure
- `TestError_MalformedFB2`: Handle malformed FB2
- `TestError_MissingElements`: Handle missing required elements
- `TestError_ConversionFailure`: Handle conversion errors
- `TestError_ErrorMessages`: Verify error messages are returned
- `TestError_ErrorStatus`: Verify error status is set

**Test Data:**
- `testdata/invalid/malformed.xml`
- `testdata/invalid/missing-elements.fb2`

### 6. File Size and Limits Tests

**Test Cases:**
- `TestFileSize_AtLimit`: Test file exactly at MAX_FILE_SIZE
- `TestFileSize_OneByteOver`: Test file one byte over limit
- `TestFileSize_RequestBodyLimit`: Test request body size limit
- `TestFileSize_MultipartLimit`: Test multipart form size limit
- `TestFileSize_ErrorMessage`: Verify error message for oversized files

### 7. Concurrency Tests

**Test Cases:**
- `TestConcurrency_MultipleConversions`: Multiple simultaneous conversions
- `TestConcurrency_StatusChecks`: Status checks during conversion
- `TestConcurrency_Downloads`: Download requests during conversion
- `TestConcurrency_Cleanup`: Cleanup during active conversions
- `TestConcurrency_JobMap`: Test job map thread safety

### 8. Temp Directory Tests

**Test Cases:**
- `TestTempDir_Creation`: Test temp directory creation
- `TestTempDir_Permissions`: Test directory permissions
- `TestTempDir_JobSubdirs`: Test job-specific subdirectory creation
- `TestTempDir_Cleanup`: Test file cleanup after conversion
- `TestTempDir_ErrorCleanup`: Test cleanup on errors

## Test Data Requirements

### Minimal Valid FB2 File

```xml
<?xml version="1.0" encoding="UTF-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
  <description>
    <title-info>
      <book-title>Test Book</book-title>
      <author>
        <first-name>Test</first-name>
        <last-name>Author</last-name>
      </author>
    </title-info>
  </description>
  <body>
    <section>
      <title>
        <p>Chapter 1</p>
      </title>
      <p>This is a test paragraph.</p>
    </section>
  </body>
</FictionBook>
```

### Complete FB2 File

Should include:
- Multiple sections with nesting
- Images (base64 encoded)
- Links
- Strong and emphasis formatting
- Poems
- Citations
- Unicode characters

## Test Implementation Guidelines

### 1. Use Table-Driven Tests

```go
func TestParseFB2(t *testing.T) {
    tests := []struct {
        name    string
        file    string
        wantErr bool
    }{
        {"valid file", "testdata/valid/minimal.fb2", false},
        {"invalid XML", "testdata/invalid/malformed.xml", true},
        {"missing file", "testdata/nonexistent.fb2", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### 2. Use Test Helpers

Create helper functions for:
- Setting up test servers
- Creating test FB2 files
- Setting up temp directories
- Cleaning up after tests

### 3. Use Mocking Where Appropriate

- Mock file system operations
- Mock HTTP requests/responses
- Use `httptest` package for HTTP testing

### 4. Test Coverage Goals

- Aim for >80% code coverage
- Focus on critical paths
- Test error handling
- Test edge cases

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Tests with Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Specific Test
```bash
go test -v ./config -run TestLoad
```

### Run Tests in Parallel
```bash
go test -parallel 4 ./...
```

## Test Execution Order

1. **Unit Tests** (fast, isolated)
   - Config tests
   - Parser tests
   - Generator tests

2. **Handler Tests** (requires mocks)
   - Converter handler tests
   - Cleanup tests

3. **Integration Tests** (slower, requires server)
   - Full conversion flow
   - API endpoint tests

4. **Edge Case Tests** (various scenarios)
   - Error handling
   - Edge cases
   - Concurrency

## Continuous Integration

Tests should be run:
- On every commit (via GitHub Actions)
- Before merging PRs
- As part of release process

## Success Criteria

Tests are considered successful when:
- All tests pass
- Code coverage >80%
- No race conditions detected
- All edge cases handled
- Error scenarios properly tested

## Next Steps

1. Create testdata directory structure
2. Generate test FB2 files
3. Implement unit tests (start with config, then parser, then generator)
4. Implement handler tests
5. Implement integration tests
6. Set up CI/CD for automated testing
7. Monitor test coverage and improve as needed

