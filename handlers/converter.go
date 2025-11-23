// Package handlers provides HTTP request handlers for the FB2 to EPUB conversion service.
package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lex/fb2epub/config"
	"github.com/lex/fb2epub/converter"
)

var conversionJobs = make(map[string]*ConversionJob)

const (
	jobStatusPending    = "pending"
	jobStatusProcessing = "processing"
	jobStatusCompleted  = "completed"
	jobStatusFailed     = "failed"
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
	if err := os.MkdirAll(cfg.TempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create base temporary directory: %v", err),
		})
		return
	}

	tempDir := filepath.Join(cfg.TempDir, jobID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create temporary directory: %v", err),
		})
		return
	}

	// Save uploaded file
	inputPath := filepath.Join(tempDir, "input.fb2")
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

func processConversion(jobID, inputPath, outputPath string, _ *config.Config) {
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
		job.Status = jobStatusFailed
		job.Error = fmt.Sprintf("Failed to parse FB2: %v", err)
		return
	}

	// Generate EPUB
	if err := converter.GenerateEPUB(fb2, outputPath); err != nil {
		job.Status = jobStatusFailed
		job.Error = fmt.Sprintf("Failed to generate EPUB: %v", err)
		return
	}

	job.Status = jobStatusCompleted
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

	if job.Status == jobStatusCompleted {
		response["download_url"] = fmt.Sprintf("/api/v1/download/%s", jobID)
	}

	if job.Status == jobStatusFailed {
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

	if job.Status != jobStatusCompleted {
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
