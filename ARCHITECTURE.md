# Service Architecture

## Overview

The FB2 to EPUB converter is a RESTful microservice built in Go that converts FictionBook 2.0 (FB2) XML files to EPUB format. The service uses asynchronous job processing to handle file conversions.

## Architecture Components

### 1. HTTP Server (Gin Framework)
- **Location**: `main.go`
- **Purpose**: Entry point, route configuration, server initialization
- **Features**:
  - Health check endpoint
  - RESTful API endpoints
  - Environment-based configuration

### 2. Configuration Management
- **Location**: `config/config.go`
- **Purpose**: Load and manage application configuration
- **Configuration Sources**:
  - Environment variables
  - Default values
- **Configurable Options**:
  - Server port
  - Environment mode (development/production)
  - Temporary directory path
  - Maximum file size limit

### 3. Data Models
- **Location**: `models/fb2.go`
- **Purpose**: Define Go structs that map to FB2 XML structure
- **Key Models**:
  - `FictionBook`: Root element
  - `Description`: Metadata (title, author, etc.)
  - `Body`: Main content
  - `Section`: Nested content sections
  - `Paragraph`: Text paragraphs with inline formatting

### 4. FB2 Parser
- **Location**: `converter/fb2parser.go`
- **Purpose**: Parse FB2 XML files into Go structs
- **Features**:
  - XML decoding with proper encoding handling
  - File and reader-based parsing
  - Error handling and validation

### 5. EPUB Generator
- **Location**: `converter/epubgenerator.go`
- **Purpose**: Generate EPUB files from parsed FB2 data
- **EPUB Structure**:
  - `mimetype`: EPUB MIME type (must be first, uncompressed)
  - `META-INF/container.xml`: Container metadata
  - `OEBPS/content.opf`: Package document (metadata, manifest, spine)
  - `OEBPS/toc.ncx`: Navigation control file
  - `OEBPS/cover.xhtml`: Cover page
  - `OEBPS/content.xhtml`: Main content
- **Features**:
  - Proper EPUB 3.0 structure
  - HTML content generation
  - Metadata extraction and formatting
  - Section and paragraph processing
  - Support for poems, citations, and formatting

### 6. HTTP Handlers
- **Location**: `handlers/converter.go`
- **Purpose**: Handle HTTP requests and responses
- **Endpoints**:
  - `POST /api/v1/convert`: Upload and convert FB2 file
  - `GET /api/v1/status/:id`: Check conversion status
  - `GET /api/v1/download/:id`: Download converted EPUB
- **Features**:
  - File upload handling
  - Asynchronous job processing
  - Job status tracking
  - File download with proper headers

## Data Flow

```
1. Client uploads FB2 file
   ↓
2. Handler receives file, creates job
   ↓
3. File saved to temp directory
   ↓
4. Async goroutine processes conversion:
   a. Parse FB2 XML → FictionBook struct
   b. Generate EPUB from struct
   c. Save EPUB to temp directory
   ↓
5. Client polls status endpoint
   ↓
6. Client downloads EPUB when ready
```

## Job Processing

- **Job States**: `pending` → `processing` → `completed` / `failed`
- **Storage**: In-memory map (for production, consider Redis/database)
- **Cleanup**: Input files deleted after processing
- **Output**: EPUB files stored temporarily (consider cleanup job)

## File Structure

```
fb2epub/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── config/
│   └── config.go          # Configuration management
├── models/
│   └── fb2.go             # FB2 data structures
├── converter/
│   ├── fb2parser.go       # FB2 XML parser
│   └── epubgenerator.go   # EPUB generator
└── handlers/
    └── converter.go       # HTTP handlers
```

## Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Dependencies**:
  - `github.com/gin-gonic/gin`: HTTP web framework
  - `github.com/google/uuid`: UUID generation
  - Standard library: `encoding/xml`, `archive/zip`, `html`

## Design Decisions

### 1. Asynchronous Processing
- **Why**: File conversion can take time, especially for large files
- **Implementation**: Goroutines for background processing
- **Trade-off**: Requires status polling (could use WebSockets for real-time)

### 2. In-Memory Job Storage
- **Why**: Simple implementation, no external dependencies
- **Limitation**: Jobs lost on restart
- **Future**: Consider Redis or database for persistence

### 3. Temporary File Storage
- **Why**: Simple file-based approach
- **Consideration**: Need cleanup mechanism for old files
- **Future**: Consider object storage (S3) for scalability

### 4. EPUB 3.0 Format
- **Why**: Modern standard, better compatibility
- **Structure**: ZIP archive with specific file layout
- **Compliance**: Follows EPUB 3.0 specification

## Scalability Considerations

### Current Limitations
- In-memory job storage (lost on restart)
- No horizontal scaling support
- No rate limiting
- No authentication/authorization

### Future Enhancements
1. **Persistent Storage**: Redis or database for jobs
2. **Queue System**: RabbitMQ/Kafka for job processing
3. **Object Storage**: S3 for file storage
4. **Caching**: Cache converted files
5. **Rate Limiting**: Prevent abuse
6. **Authentication**: API keys or OAuth
7. **Monitoring**: Metrics and logging
8. **Horizontal Scaling**: Load balancer + multiple instances

## Security Considerations

1. **File Size Limits**: Prevent DoS via large files
2. **File Type Validation**: Only accept .fb2/.xml files
3. **Path Traversal**: Sanitize file paths
4. **Resource Limits**: Memory and CPU limits
5. **Input Validation**: Validate XML structure
6. **HTTPS**: Use TLS in production
7. **Rate Limiting**: Prevent abuse (future)

## Testing Strategy

### Unit Tests
- FB2 parser with various XML structures
- EPUB generator with different content types
- Handler logic

### Integration Tests
- End-to-end conversion workflow
- Error handling scenarios
- File upload/download

### Load Tests
- Concurrent conversions
- Large file handling
- Memory usage

## Deployment Architecture

### Development
- Single binary
- Local file system
- In-memory storage

### Production
- Systemd service
- Nginx reverse proxy
- SSL/TLS
- Log rotation
- Monitoring

## Monitoring Points

1. **Request Rate**: Requests per second
2. **Conversion Time**: Average processing time
3. **Success Rate**: Successful vs failed conversions
4. **File Sizes**: Average input/output sizes
5. **Error Rates**: Error types and frequencies
6. **Resource Usage**: CPU, memory, disk

## Error Handling

- **File Upload Errors**: Invalid format, too large
- **Parsing Errors**: Malformed XML, encoding issues
- **Generation Errors**: EPUB structure issues
- **Storage Errors**: Disk full, permission issues
- **Network Errors**: Timeout, connection issues

All errors are logged and returned to client with appropriate HTTP status codes.

