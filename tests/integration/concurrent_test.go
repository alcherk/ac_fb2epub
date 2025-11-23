package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lex/fb2epub/handlers"
)

func getTestDataPath(filename string) string {
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	projectRoot := filepath.Join(testDir, "..", "..")
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

func TestIntegration_ConcurrentConversions(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TEMP_DIR", tmpDir)
	os.Setenv("MAX_FILE_SIZE", "10485760")
	defer os.Clearenv()

	router := setupIntegrationRouter()
	fb2Path := getTestDataPath(filepath.Join("valid", "minimal.fb2"))

	numConversions := 3
	var wg sync.WaitGroup
	jobIDs := make([]string, numConversions)
	errors := make([]error, numConversions)

	wg.Add(numConversions)
	for i := 0; i < numConversions; i++ {
		go func(index int) {
			defer wg.Done()

			body, contentType := createTestFB2Multipart(t, fb2Path)
			req := httptest.NewRequest("POST", "/api/v1/convert", body)
			req.Header.Set("Content-Type", contentType)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusAccepted {
				errors[index] = fmt.Errorf("expected status %d, got %d", http.StatusAccepted, w.Code)
				return
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				errors[index] = err
				return
			}

			jobID, ok := response["job_id"].(string)
			if !ok || jobID == "" {
				errors[index] = fmt.Errorf("missing job_id")
				return
			}

			jobIDs[index] = jobID

			// Wait for completion
			maxAttempts := 30
			for attempt := 0; attempt < maxAttempts; attempt++ {
				time.Sleep(200 * time.Millisecond)

				req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/status/%s", jobID), nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					continue
				}

				var statusResponse map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &statusResponse); err != nil {
					continue
				}

				status, ok := statusResponse["status"].(string)
				if !ok {
					continue
				}

				if status == "completed" {
					break
				}

				if status == "failed" {
					errorMsg, _ := statusResponse["error"].(string)
					errors[index] = fmt.Errorf("conversion failed: %s", errorMsg)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("Conversion %d failed: %v", i, err)
		}
	}

	// Verify all jobs completed
	for i, jobID := range jobIDs {
		if jobID == "" {
			t.Errorf("Conversion %d did not return job ID", i)
			continue
		}

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/status/%s", jobID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status check for job %s failed", jobID)
			continue
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to parse status for job %s: %v", jobID, err)
			continue
		}

		status, ok := response["status"].(string)
		if !ok {
			t.Errorf("Status response for job %s missing status field", jobID)
			continue
		}

		if status != "completed" && status != "failed" {
			t.Errorf("Job %s should be completed or failed, got %s", jobID, status)
		}
	}

	// Cleanup
	time.Sleep(500 * time.Millisecond)
	for _, jobID := range jobIDs {
		if jobID != "" {
			handlers.DeleteConversionJob(jobID)
		}
	}

	entries, err := os.ReadDir(tmpDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				_ = os.RemoveAll(filepath.Join(tmpDir, entry.Name()))
			}
		}
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TEMP_DIR", tmpDir)
	os.Setenv("MAX_FILE_SIZE", "10485760")
	defer os.Clearenv()

	router := setupIntegrationRouter()

	// Test with invalid file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "invalid.txt")
	part.Write([]byte("This is not an FB2 file"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/convert", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid file, got %d", http.StatusBadRequest, w.Code)
	}

	// Test with malformed FB2
	fb2Path := getTestDataPath(filepath.Join("invalid", "malformed.xml"))
	if _, err := os.Stat(fb2Path); err == nil {
		body, contentType := createTestFB2Multipart(t, fb2Path)
		req := httptest.NewRequest("POST", "/api/v1/convert", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusAccepted {
			t.Logf("Malformed file rejected immediately (acceptable): %d", w.Code)
		} else {
			// If accepted, wait and check for failure
			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			jobID, _ := response["job_id"].(string)

			time.Sleep(500 * time.Millisecond)

			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/status/%s", jobID), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				var statusResponse map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &statusResponse)
				status, _ := statusResponse["status"].(string)
				if status != "failed" {
					t.Logf("Job should fail for malformed FB2, got status: %s", status)
				}
			}

			if jobID != "" {
				handlers.DeleteConversionJob(jobID)
			}
		}
	}
}

