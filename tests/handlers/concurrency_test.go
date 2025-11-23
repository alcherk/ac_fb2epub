package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/lex/fb2epub/handlers"
)

func TestConcurrency_MultipleConversions(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TEMP_DIR", tmpDir)
	os.Setenv("MAX_FILE_SIZE", "10485760") // 10MB
	defer os.Clearenv()

	router := setupTestRouter()
	
	// Number of concurrent conversions
	numConversions := 5
	var wg sync.WaitGroup
	jobIDs := make([]string, numConversions)
	errors := make([]error, numConversions)
	
	wg.Add(numConversions)
	for i := 0; i < numConversions; i++ {
		go func(index int) {
			defer wg.Done()
			
			body, contentType := createTestFB2File(t)
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
				errors[index] = fmt.Errorf("missing job_id in response")
				return
			}
			
			jobIDs[index] = jobID
		}(i)
	}
	
	wg.Wait()
	
	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("Conversion %d failed: %v", i, err)
		}
	}
	
	// Verify all jobs were created
	for i, jobID := range jobIDs {
		if jobID == "" {
			t.Errorf("Conversion %d did not return job ID", i)
			continue
		}
		
		job := handlers.GetConversionJob(jobID)
		if job == nil {
			t.Errorf("Job %s was not created", jobID)
		}
	}
	
	// Cleanup
	time.Sleep(500 * time.Millisecond) // Wait for processing
	for _, jobID := range jobIDs {
		if jobID != "" {
			handlers.DeleteConversionJob(jobID)
		}
	}
	
	// Clean up directories
	entries, err := os.ReadDir(tmpDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				_ = os.RemoveAll(filepath.Join(tmpDir, entry.Name()))
			}
		}
	}
}

func TestConcurrency_StatusChecks(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TEMP_DIR", tmpDir)
	os.Setenv("MAX_FILE_SIZE", "10485760")
	defer os.Clearenv()

	router := setupTestRouter()
	body, contentType := createTestFB2File(t)

	// Create a job
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

	jobID, ok := convertResponse["job_id"].(string)
	if !ok {
		t.Fatal("Response should contain job_id")
	}
	defer handlers.DeleteConversionJob(jobID)

	// Concurrent status checks
	numChecks := 10
	var wg sync.WaitGroup
	errors := make([]error, numChecks)
	
	wg.Add(numChecks)
	for i := 0; i < numChecks; i++ {
		go func(index int) {
			defer wg.Done()
			
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/status/%s", jobID), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				errors[index] = fmt.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
				return
			}
			
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				errors[index] = err
				return
			}
			
			if response["id"] != jobID {
				errors[index] = fmt.Errorf("expected job ID %s, got %v", jobID, response["id"])
			}
		}(i)
	}
	
	wg.Wait()
	
	// Wait for async processing and cleanup
	time.Sleep(500 * time.Millisecond)
	
	// Clean up any job directories
	entries, err := os.ReadDir(tmpDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				_ = os.RemoveAll(filepath.Join(tmpDir, entry.Name()))
			}
		}
	}
	
	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("Status check %d failed: %v", i, err)
		}
	}
}

func TestConcurrency_Downloads(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TEMP_DIR", tmpDir)
	defer os.Clearenv()

	// Create a completed job
	jobID := "concurrent-download-job-id"
	jobDir := filepath.Join(tmpDir, jobID)
	os.MkdirAll(jobDir, 0755)
	
	epubPath := filepath.Join(jobDir, "output.epub")
	file, err := os.Create(epubPath)
	if err != nil {
		t.Fatalf("Failed to create test EPUB: %v", err)
	}
	file.WriteString("EPUB content")
	file.Close()

	job := &handlers.ConversionJob{
		ID:        jobID,
		Status:    handlers.JobStatusCompleted,
		CreatedAt: time.Now(),
		FilePath:  epubPath,
	}
	handlers.SetConversionJob(job)
	defer handlers.DeleteConversionJob(jobID)

	router := setupTestRouter()
	
	// Concurrent download requests
	numDownloads := 5
	var wg sync.WaitGroup
	errors := make([]error, numDownloads)
	
	wg.Add(numDownloads)
	for i := 0; i < numDownloads; i++ {
		go func(index int) {
			defer wg.Done()
			
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/download/%s", jobID), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				errors[index] = fmt.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
				return
			}
			
			if w.Body.Len() == 0 {
				errors[index] = fmt.Errorf("download body is empty")
			}
		}(i)
	}
	
	wg.Wait()
	
	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("Download %d failed: %v", i, err)
		}
	}
}

func TestConcurrency_JobMap(t *testing.T) {
	// Test concurrent access to job map
	numJobs := 5
	var wg sync.WaitGroup
	
	// Concurrent job creation
	wg.Add(numJobs)
	for i := 0; i < numJobs; i++ {
		go func(index int) {
			defer wg.Done()
			
			jobID := fmt.Sprintf("job-%d", index)
			job := &handlers.ConversionJob{
				ID:        jobID,
				Status:    handlers.JobStatusProcessing,
				CreatedAt: time.Now(),
				FilePath:  fmt.Sprintf("/tmp/%s.epub", jobID),
			}
			handlers.SetConversionJob(job)
		}(i)
	}
	
	wg.Wait() // Wait for all writes to complete
	
	// Verify all jobs were created (sequential check to avoid race conditions)
	for i := 0; i < numJobs; i++ {
		jobID := fmt.Sprintf("job-%d", i)
		job := handlers.GetConversionJob(jobID)
		if job == nil {
			t.Errorf("Job %s not found after creation", jobID)
			continue
		}
		if job.ID != jobID {
			t.Errorf("Expected job ID %s, got %s", jobID, job.ID)
		}
	}
	
	// Concurrent reads (jobs already exist)
	wg.Add(numJobs)
	for i := 0; i < numJobs; i++ {
		go func(index int) {
			defer wg.Done()
			
			jobID := fmt.Sprintf("job-%d", index)
			job := handlers.GetConversionJob(jobID)
			if job == nil {
				t.Errorf("Job %s not found during concurrent read", jobID)
				return
			}
			if job.ID != jobID {
				t.Errorf("Expected job ID %s, got %s", jobID, job.ID)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Cleanup
	for i := 0; i < numJobs; i++ {
		jobID := fmt.Sprintf("job-%d", i)
		handlers.DeleteConversionJob(jobID)
	}
}

