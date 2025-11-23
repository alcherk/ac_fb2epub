// Package handlers provides HTTP request handlers for the FB2 to EPUB conversion service.
package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lex/fb2epub/config"
	"github.com/lex/fb2epub/converter"
)

var (
	conversionJobs    = make(map[string]*ConversionJob)
	completedJobCount = 0        // Counter for completed conversions
	cleanupMutex      sync.Mutex // Mutex for cleanup operations
)

// Job status constants
const (
	JobStatusPending    = "pending"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
)

// ConversionJob represents a file conversion job
type ConversionJob struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"` // pending, processing, completed, failed
	CreatedAt time.Time `json:"created_at"`
	FilePath  string    `json:"-"`
	Error     string    `json:"error,omitempty"`
}

// ConvertFB2ToEPUB handles the conversion request
func ConvertFB2ToEPUB(c *gin.Context) {
	cfg := config.Load()

	// Check file size - set MaxBytesReader with a buffer to handle large files
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.MaxFileSize)

	// Parse multipart form with increased size limit
	// Note: This must be set before parsing
	if err := c.Request.ParseMultipartForm(cfg.MaxFileSize); err != nil {
		// Check if it's a size-related error
		if err.Error() == "http: request body too large" ||
			err.Error() == "multipart: NextPart: EOF" ||
			err.Error() == "http: request body too large" {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("File too large. Maximum size: %d bytes (%.2f MB)",
					cfg.MaxFileSize, float64(cfg.MaxFileSize)/(1024*1024)),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Failed to parse form data: %v", err),
			})
		}
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file provided or invalid file",
		})
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	// Validate file extension
	ext := filepath.Ext(header.Filename)
	if ext != ".fb2" && ext != ".xml" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file type. Expected .fb2 or .xml file",
		})
		return
	}

	// Create job ID
	jobID := uuid.New().String()

	// Create temp directory for this job
	// Ensure base temp directory exists first
	//nolint:gosec // 0755 needed for Docker volume mounts
	if err := os.MkdirAll(cfg.TempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create base temporary directory: %v", err),
		})
		return
	}

	tempDir := filepath.Join(cfg.TempDir, jobID)
	//nolint:gosec // 0755 needed for Docker volume mounts
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create temporary directory: %v", err),
		})
		return
	}

	// Save uploaded file
	inputPath := filepath.Join(tempDir, "input.fb2")
	//nolint:gosec // Path is controlled and validated
	outFile, err := os.Create(inputPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save uploaded file",
		})
		return
	}

	_, err = io.Copy(outFile, file)
	if closeErr := outFile.Close(); closeErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save uploaded file",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save uploaded file",
		})
		return
	}

	// Create job
	job := &ConversionJob{
		ID:        jobID,
		Status:    "processing",
		CreatedAt: time.Now(),
		FilePath:  filepath.Join(tempDir, "output.epub"),
	}
	conversionJobs[jobID] = job

	// Process conversion asynchronously
	go processConversion(jobID, inputPath, job.FilePath, cfg)

	// Return job ID immediately
	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"status":  "processing",
		"message": "Conversion started",
	})
}

func processConversion(jobID, inputPath, outputPath string, cfg *config.Config) {
	job := conversionJobs[jobID]
	defer func() {
		// Cleanup input file after processing
		if removeErr := os.Remove(inputPath); removeErr != nil {
			_ = removeErr
		}
	}()

	// Parse FB2
	fb2, err := converter.ParseFB2(inputPath)
	if err != nil {
		job.Status = JobStatusFailed
		job.Error = fmt.Sprintf("Failed to parse FB2: %v", err)
		return
	}

	// Generate EPUB
	if err := converter.GenerateEPUB(fb2, outputPath); err != nil {
		job.Status = JobStatusFailed
		job.Error = fmt.Sprintf("Failed to generate EPUB: %v", err)
		return
	}

	job.Status = JobStatusCompleted

	// Increment completed job counter and trigger cleanup if needed
	cleanupMutex.Lock()
	completedJobCount++
	shouldCleanup := completedJobCount >= cfg.CleanupTriggerCount
	if shouldCleanup {
		completedJobCount = 0 // Reset counter
	}
	cleanupMutex.Unlock()

	if shouldCleanup {
		// Trigger cleanup asynchronously
		go cleanupOldJobs(cfg)
	}
}

// GetConversionStatus returns the status of a conversion job
func GetConversionStatus(c *gin.Context) {
	jobID := c.Param("id")

	job, exists := conversionJobs[jobID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	response := gin.H{
		"id":         job.ID,
		"status":     job.Status,
		"created_at": job.CreatedAt,
	}

	if job.Status == JobStatusCompleted {
		response["download_url"] = fmt.Sprintf("/api/v1/download/%s", jobID)
	}

	if job.Status == JobStatusFailed {
		response["error"] = job.Error
	}

	c.JSON(http.StatusOK, response)
}

// DownloadEPUB handles EPUB file download
func DownloadEPUB(c *gin.Context) {
	jobID := c.Param("id")

	job, exists := conversionJobs[jobID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	if job.Status != JobStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Conversion not completed yet",
		})
		return
	}

	// Check if file exists
	if _, err := os.Stat(job.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "EPUB file not found",
		})
		return
	}

	// Set headers for file download
	c.Header("Content-Type", "application/epub+zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"book_%s.epub\"", jobID))

	// Send file
	c.File(job.FilePath)
}

// cleanupOldJobs removes old job directories from the temp folder
func cleanupOldJobs(cfg *config.Config) {
	// Use mutex to prevent concurrent cleanup operations
	cleanupMutex.Lock()
	defer cleanupMutex.Unlock()

	// Get all directories in temp folder
	entries, err := os.ReadDir(cfg.TempDir)
	if err != nil {
		return
	}

	now := time.Now()
	cleanedCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if it's a valid UUID (job ID format)
		jobID := entry.Name()
		if len(jobID) != 36 { // UUIDs are 36 characters
			continue
		}

		// Get job info
		job, exists := conversionJobs[jobID]
		jobDir := filepath.Join(cfg.TempDir, jobID)

		// Cleanup conditions:
		// 1. Job doesn't exist in memory (old job) and directory is older than 1 hour
		// 2. Job is completed and older than 1 hour
		// 3. Job is failed and older than 1 hour
		shouldCleanup := false
		if !exists {
			// Job not in memory, check directory age
			info, err := os.Stat(jobDir)
			if err == nil {
				if now.Sub(info.ModTime()) > time.Hour {
					shouldCleanup = true
				}
			}
		} else if job.Status == JobStatusCompleted || job.Status == JobStatusFailed {
			// Job is completed or failed, check if older than 1 hour
			if now.Sub(job.CreatedAt) > time.Hour {
				shouldCleanup = true
			}
		}

		if shouldCleanup {
			// Remove the entire job directory
			if err := os.RemoveAll(jobDir); err == nil {
				cleanedCount++
				// Remove from memory if exists
				if exists {
					delete(conversionJobs, jobID)
				}
			}
		}
	}

	// Log cleanup result (in production, you might want to use a proper logger)
	if cleanedCount > 0 {
		// Suppress unused variable warning - in production, log this
		_ = cleanedCount
	}
}

// GetConversionJob returns a conversion job by ID (for testing)
func GetConversionJob(jobID string) *ConversionJob {
	return conversionJobs[jobID]
}

// SetConversionJob sets a conversion job (for testing)
func SetConversionJob(job *ConversionJob) {
	conversionJobs[job.ID] = job
}

// DeleteConversionJob deletes a conversion job (for testing)
func DeleteConversionJob(jobID string) {
	delete(conversionJobs, jobID)
}
