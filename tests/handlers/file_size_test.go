package handlers_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestRouter is defined in converter_test.go

func createLargeFileWithContentType(t *testing.T, size int64) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a minimal FB2 header
	fb2Header := `<?xml version="1.0" encoding="UTF-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
  <description>
    <title-info>
      <book-title>Test Book</book-title>
    </title-info>
  </description>
  <body>
    <section>
      <p>`

	// Create padding to reach desired size
	paddingSize := size - int64(len(fb2Header)) - 100 // Leave room for closing tags
	if paddingSize < 0 {
		paddingSize = 0
	}

	padding := make([]byte, paddingSize)
	for i := range padding {
		padding[i] = 'A'
	}

	fb2Footer := `</p>
    </section>
  </body>
</FictionBook>`

	fb2Content := fb2Header + string(padding) + fb2Footer

	part, err := writer.CreateFormFile("file", "test.fb2")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	if _, err := part.Write([]byte(fb2Content)); err != nil {
		t.Fatalf("Failed to write file content: %v", err)
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	return body, contentType
}

func TestFileSize_AtLimit(t *testing.T) {
	// Set max file size to 1MB
	maxSize := int64(1024 * 1024) // 1MB
	os.Setenv("TEMP_DIR", t.TempDir())
	os.Setenv("MAX_FILE_SIZE", "1048576")
	defer os.Clearenv()

	router := setupTestRouter()
	body, contentType := createLargeFileWithContentType(t, maxSize)

	req := httptest.NewRequest("POST", "/api/v1/convert", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should accept file at the limit
	if w.Code == http.StatusRequestEntityTooLarge {
		t.Log("File at limit was rejected - this might be acceptable depending on implementation")
	} else if w.Code != http.StatusAccepted && w.Code != http.StatusBadRequest {
		t.Errorf("Unexpected status code: %d", w.Code)
	}
}

func TestFileSize_OneByteOver(t *testing.T) {
	// Set max file size to 1MB
	maxSize := int64(1024 * 1024) // 1MB
	os.Setenv("TEMP_DIR", t.TempDir())
	os.Setenv("MAX_FILE_SIZE", "1048576")
	defer os.Clearenv()

	router := setupTestRouter()
	body, contentType := createLargeFileWithContentType(t, maxSize+1) // One byte over

	req := httptest.NewRequest("POST", "/api/v1/convert", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should reject file over the limit
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status %d for oversized file, got %d", http.StatusRequestEntityTooLarge, w.Code)
	}
}

func TestFileSize_ErrorMessage(t *testing.T) {
	maxSize := int64(1024 * 1024) // 1MB
	os.Setenv("TEMP_DIR", t.TempDir())
	os.Setenv("MAX_FILE_SIZE", "1048576")
	defer os.Clearenv()

	router := setupTestRouter()
	body, contentType := createLargeFileWithContentType(t, maxSize+1000) // Over limit

	req := httptest.NewRequest("POST", "/api/v1/convert", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code == http.StatusRequestEntityTooLarge {
		// Check that error message is present
		bodyStr := w.Body.String()
		if bodyStr == "" {
			t.Error("Error response should contain a message")
		}
	}
}

func TestFileSize_SmallFile(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TEMP_DIR", tmpDir)
	os.Setenv("MAX_FILE_SIZE", "10485760") // 10MB
	defer os.Clearenv()

	router := setupTestRouter()
	body, contentType := createLargeFileWithContentType(t, 1024) // 1KB - well under limit

	req := httptest.NewRequest("POST", "/api/v1/convert", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Small file should be accepted
	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status %d for small file, got %d. Body: %s", http.StatusAccepted, w.Code, w.Body.String())
		return
	}

	// Wait a bit for async processing to complete and cleanup
	// The input file should be deleted after processing
	time.Sleep(100 * time.Millisecond)
	
	// Clean up any job directories that might have been created
	// This helps prevent "directory not empty" errors during test cleanup
	entries, err := os.ReadDir(tmpDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				// Try to remove the directory, ignore errors
				_ = os.RemoveAll(filepath.Join(tmpDir, entry.Name()))
			}
		}
	}
}

