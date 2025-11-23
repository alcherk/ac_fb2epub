package tests_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lex/fb2epub/handlers"
)

func getTestDataPath(filename string) string {
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	projectRoot := filepath.Join(testDir, "..")
	return filepath.Join(projectRoot, "testdata", filename)
}

func setupIntegrationRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/convert", handlers.ConvertFB2ToEPUB)
	router.GET("/api/v1/status/:id", handlers.GetConversionStatus)
	router.GET("/api/v1/download/:id", handlers.DownloadEPUB)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "fb2epub",
		})
	})
	return router
}

func createTestFB2Multipart(t *testing.T, fb2Path string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Read the FB2 file
	fb2Content, err := os.ReadFile(fb2Path)
	if err != nil {
		t.Fatalf("Failed to read test FB2 file: %v", err)
	}

	part, err := writer.CreateFormFile("file", filepath.Base(fb2Path))
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	if _, err := part.Write(fb2Content); err != nil {
		t.Fatalf("Failed to write file content: %v", err)
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	return body, contentType
}

func TestIntegration_FullConversion(t *testing.T) {
	// Set up test environment
	tmpDir := t.TempDir()
	os.Setenv("TEMP_DIR", tmpDir)
	os.Setenv("MAX_FILE_SIZE", "10485760") // 10MB
	defer os.Clearenv()

	router := setupIntegrationRouter()
	fb2Path := getTestDataPath(filepath.Join("valid", "minimal.fb2"))
	body, contentType := createTestFB2Multipart(t, fb2Path)

	// Step 1: Upload file
	req := httptest.NewRequest("POST", "/api/v1/convert", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusAccepted, w.Code, w.Body.String())
	}

	var convertResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &convertResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	jobID, ok := convertResponse["job_id"].(string)
	if !ok {
		t.Fatal("Response should contain job_id")
	}

	// Step 2: Poll status until completed
	maxAttempts := 30
	var statusResponse map[string]interface{}

	for i := 0; i < maxAttempts; i++ {
		time.Sleep(100 * time.Millisecond) // Small delay between polls

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/status/%s", jobID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Status check failed with code %d", w.Code)
		}

		if err := json.Unmarshal(w.Body.Bytes(), &statusResponse); err != nil {
			t.Fatalf("Failed to parse status response: %v", err)
		}

		status, ok := statusResponse["status"].(string)
		if !ok {
			t.Fatal("Status response should contain status field")
		}

		if status == "completed" {
			break
		}

		if status == "failed" {
			errorMsg, _ := statusResponse["error"].(string)
			t.Fatalf("Conversion failed: %s", errorMsg)
		}

		if i == maxAttempts-1 {
			t.Fatal("Conversion did not complete within timeout")
		}
	}

	// Step 3: Download EPUB
	downloadURL, ok := statusResponse["download_url"].(string)
	if !ok {
		t.Fatal("Completed job should have download_url")
	}

	req = httptest.NewRequest("GET", downloadURL, nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Download failed with code %d", w.Code)
	}

	// Step 4: Verify EPUB is valid
	epubData := w.Body.Bytes()
	if len(epubData) == 0 {
		t.Fatal("Downloaded EPUB is empty")
	}

	// Save to temp file for validation
	epubPath := filepath.Join(tmpDir, "downloaded.epub")
	if err := os.WriteFile(epubPath, epubData, 0644); err != nil {
		t.Fatalf("Failed to save EPUB: %v", err)
	}

	// Verify it's a valid ZIP archive
	reader, err := zip.OpenReader(epubPath)
	if err != nil {
		t.Fatalf("Downloaded file is not a valid EPUB (ZIP): %v", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			t.Logf("Error closing ZIP: %v", closeErr)
		}
	}()

	// Check for required files
	requiredFiles := []string{
		"mimetype",
		"META-INF/container.xml",
		"OEBPS/content.opf",
	}

	for _, requiredFile := range requiredFiles {
		found := false
		for _, file := range reader.File {
			if file.Name == requiredFile {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required file %s not found in EPUB", requiredFile)
		}
	}
}

func TestIntegration_StatusPolling(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TEMP_DIR", tmpDir)
	os.Setenv("MAX_FILE_SIZE", "10485760")
	defer os.Clearenv()

	router := setupIntegrationRouter()
	fb2Path := getTestDataPath(filepath.Join("valid", "minimal.fb2"))
	body, contentType := createTestFB2Multipart(t, fb2Path)

	// Upload file
	req := httptest.NewRequest("POST", "/api/v1/convert", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("Expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	var convertResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &convertResponse); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	jobID := convertResponse["job_id"].(string)

	// Poll status multiple times
	statuses := make([]string, 0)
	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond)

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/status/%s", jobID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Status check failed with code %d", w.Code)
		}

		var statusResponse map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &statusResponse); err != nil {
			t.Fatalf("Failed to parse status response: %v", err)
		}

		status := statusResponse["status"].(string)
		statuses = append(statuses, status)

		if status == "completed" || status == "failed" {
			break
		}
	}

	// Verify we got status updates
	if len(statuses) == 0 {
		t.Error("Should have received at least one status update")
	}

	// Final status should be completed or failed
	finalStatus := statuses[len(statuses)-1]
	if finalStatus != "completed" && finalStatus != "failed" {
		t.Errorf("Final status should be completed or failed, got %s", finalStatus)
	}
}

func TestIntegration_HealthEndpoint(t *testing.T) {
	router := setupIntegrationRouter()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", response["status"])
	}

	if response["service"] != "fb2epub" {
		t.Errorf("Expected service 'fb2epub', got %v", response["service"])
	}
}

func TestIntegration_InvalidRoute(t *testing.T) {
	router := setupIntegrationRouter()

	req := httptest.NewRequest("GET", "/api/v1/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

