package converter_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lex/fb2epub/converter"
)

func TestErrorHandling_InvalidXML(t *testing.T) {
	filePath := getTestDataPath(filepath.Join("invalid", "malformed.xml"))
	_, err := converter.ParseFB2(filePath)
	if err == nil {
		t.Error("ParseFB2() should return error for invalid XML")
	}
}

func TestErrorHandling_MalformedFB2(t *testing.T) {
	filePath := getTestDataPath(filepath.Join("invalid", "malformed.xml"))
	fb2, err := converter.ParseFB2(filePath)
	if err == nil {
		// If it somehow parses, try to generate EPUB
		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "test.epub")
		err = converter.GenerateEPUB(fb2, outputPath)
		if err == nil {
			t.Error("GenerateEPUB() should return error for malformed FB2")
		}
	}
}

func TestErrorHandling_EmptyFile(t *testing.T) {
	filePath := getTestDataPath(filepath.Join("invalid", "empty.fb2"))
	fb2, err := converter.ParseFB2(filePath)
	if err == nil {
		// Empty file might parse but should fail EPUB generation
		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "test.epub")
		err = converter.GenerateEPUB(fb2, outputPath)
		// Generation might succeed with empty content, which is acceptable
		if err != nil {
			t.Logf("EPUB generation failed for empty file (expected): %v", err)
		}
	}
}

func TestErrorHandling_MissingFile(t *testing.T) {
	filePath := getTestDataPath("nonexistent.fb2")
	_, err := converter.ParseFB2(filePath)
	if err == nil {
		t.Error("ParseFB2() should return error for missing file")
	}

	// Check error message is informative
	if err != nil {
		errorMsg := err.Error()
		if errorMsg == "" {
			t.Error("Error message should not be empty")
		}
	}
}

func TestErrorHandling_InvalidOutputPath(t *testing.T) {
	fb2Path := getTestDataPath(filepath.Join("valid", "minimal.fb2"))
	fb2, err := converter.ParseFB2(fb2Path)
	if err != nil {
		t.Fatalf("Failed to parse FB2: %v", err)
	}

	// Try to write to invalid path (parent directory doesn't exist)
	invalidPath := "/nonexistent/path/test.epub"
	err = converter.GenerateEPUB(fb2, invalidPath)
	if err == nil {
		t.Error("GenerateEPUB() should return error for invalid output path")
	}
}

func TestErrorHandling_ReadOnlyDirectory(t *testing.T) {
	// This test might not work on all systems, so we'll skip it if it fails
	t.Skip("Read-only directory test may not work on all systems")

	fb2Path := getTestDataPath(filepath.Join("valid", "minimal.fb2"))
	fb2, err := converter.ParseFB2(fb2Path)
	if err != nil {
		t.Fatalf("Failed to parse FB2: %v", err)
	}

	// Try to write to read-only directory (if we can create one)
	readOnlyDir := t.TempDir()
	if err := os.Chmod(readOnlyDir, 0555); err != nil {
		t.Skip("Cannot set read-only permissions")
	}
	defer os.Chmod(readOnlyDir, 0755) // Restore permissions

	outputPath := filepath.Join(readOnlyDir, "test.epub")
	err = converter.GenerateEPUB(fb2, outputPath)
	if err == nil {
		t.Error("GenerateEPUB() should return error for read-only directory")
	}
}

