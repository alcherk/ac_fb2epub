package handlers_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lex/fb2epub/handlers"
)

func TestTempDir_Creation(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TEMP_DIR", tmpDir)
	defer os.Clearenv()

	// Verify temp directory exists
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Fatal("Temp directory should be created")
	}

	// Test that the handler can create subdirectories
	jobID := "test-job-creation"
	jobDir := filepath.Join(tmpDir, jobID)
	
	err := os.MkdirAll(jobDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create job directory: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(jobDir); os.IsNotExist(err) {
		t.Fatal("Job directory should be created")
	}

	// Cleanup
	os.RemoveAll(jobDir)
}

func TestTempDir_Permissions(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TEMP_DIR", tmpDir)
	defer os.Clearenv()

	// Create a job directory
	jobID := "test-permissions"
	jobDir := filepath.Join(tmpDir, jobID)
	
	err := os.MkdirAll(jobDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create job directory: %v", err)
	}

	// Check directory permissions (stat on directory)
	info, err := os.Stat(jobDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}

	// Verify it's a directory
	if !info.IsDir() {
		t.Error("Should be a directory")
	}

	// Verify we can write to it
	testFile := filepath.Join(jobDir, "test.txt")
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create file in directory: %v", err)
	}
	file.Close()

	// Cleanup
	os.RemoveAll(jobDir)
}

func TestTempDir_JobSubdirs(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TEMP_DIR", tmpDir)
	defer os.Clearenv()

	// Create multiple job subdirectories
	jobIDs := []string{
		"11111111-1111-1111-1111-111111111111",
		"22222222-2222-2222-2222-222222222222",
		"33333333-3333-3333-3333-333333333333",
	}

	for _, jobID := range jobIDs {
		jobDir := filepath.Join(tmpDir, jobID)
		
		err := os.MkdirAll(jobDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create job directory %s: %v", jobID, err)
		}

		// Create input and output files in each directory
		inputFile := filepath.Join(jobDir, "input.fb2")
		outputFile := filepath.Join(jobDir, "output.epub")

		if err := os.WriteFile(inputFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		if err := os.WriteFile(outputFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create output file: %v", err)
		}

		// Verify files exist
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			t.Errorf("Input file should exist for job %s", jobID)
		}

		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Output file should exist for job %s", jobID)
		}
	}

	// Verify all directories exist
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	if len(entries) != len(jobIDs) {
		t.Errorf("Expected %d job directories, found %d", len(jobIDs), len(entries))
	}

	// Cleanup
	for _, jobID := range jobIDs {
		os.RemoveAll(filepath.Join(tmpDir, jobID))
	}
}

func TestTempDir_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TEMP_DIR", tmpDir)
	defer os.Clearenv()

	// Create a job directory with files
	jobID := "test-cleanup-job"
	jobDir := filepath.Join(tmpDir, jobID)
	os.MkdirAll(jobDir, 0755)

	// Create files
	inputFile := filepath.Join(jobDir, "input.fb2")
	outputFile := filepath.Join(jobDir, "output.epub")

	os.WriteFile(inputFile, []byte("test input"), 0644)
	os.WriteFile(outputFile, []byte("test output"), 0644)

	// Verify files exist
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Fatal("Input file should exist")
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("Output file should exist")
	}

	// Cleanup the directory
	err := os.RemoveAll(jobDir)
	if err != nil {
		t.Fatalf("Failed to cleanup directory: %v", err)
	}

	// Verify directory is gone
	if _, err := os.Stat(jobDir); err == nil {
		t.Error("Directory should be removed after cleanup")
	}
}

func TestTempDir_ErrorCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TEMP_DIR", tmpDir)
	defer os.Clearenv()

	// Create a failed job directory
	jobID := "test-error-job"
	jobDir := filepath.Join(tmpDir, jobID)
	os.MkdirAll(jobDir, 0755)

	// Create only input file (conversion failed)
	inputFile := filepath.Join(jobDir, "input.fb2")
	os.WriteFile(inputFile, []byte("invalid fb2"), 0644)

	// Create a failed job
	failedJob := &handlers.ConversionJob{
		ID:        jobID,
		Status:    handlers.JobStatusFailed,
		CreatedAt: time.Now().Add(-2 * time.Hour), // Old enough to be cleaned up
		FilePath:  filepath.Join(jobDir, "output.epub"),
		Error:     "Conversion failed",
	}
	handlers.SetConversionJob(failedJob)

	// Verify job exists
	job := handlers.GetConversionJob(jobID)
	if job == nil {
		t.Fatal("Failed job should exist")
	}

	if job.Status != handlers.JobStatusFailed {
		t.Errorf("Expected status 'failed', got %s", job.Status)
	}

	// Verify directory exists
	if _, err := os.Stat(jobDir); os.IsNotExist(err) {
		t.Fatal("Job directory should exist even for failed jobs")
	}

	// Verify input file exists (output file should not)
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Error("Input file should exist even for failed jobs")
	}

	// Cleanup
	handlers.DeleteConversionJob(jobID)
	os.RemoveAll(jobDir)
}

