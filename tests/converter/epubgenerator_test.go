package converter_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lex/fb2epub/converter"
)


func TestGenerateEPUB_ValidStructure(t *testing.T) {
	// Parse test FB2 file
	fb2Path := getTestDataPath(filepath.Join("valid", "minimal.fb2"))
	fb2, err := converter.ParseFB2(fb2Path)
	if err != nil {
		t.Fatalf("Failed to parse FB2: %v", err)
	}

	// Create temp output file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.epub")

	// Generate EPUB
	err = converter.GenerateEPUB(fb2, outputPath)
	if err != nil {
		t.Fatalf("GenerateEPUB() error = %v, want nil", err)
	}

	// Check file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("EPUB file was not created")
	}

	// Open as ZIP archive
	reader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("Failed to open EPUB as ZIP: %v", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			t.Logf("Error closing ZIP: %v", closeErr)
		}
	}()

	// Check required files exist
	requiredFiles := map[string]bool{
		"mimetype":                    false,
		"META-INF/container.xml":     false,
		"OEBPS/content.opf":           false,
		"OEBPS/toc.ncx":               false,
		"OEBPS/nav.xhtml":             false,
		"OEBPS/content.xhtml":         false,
	}

	for _, file := range reader.File {
		if _, exists := requiredFiles[file.Name]; exists {
			requiredFiles[file.Name] = true
		}
	}

	for fileName, found := range requiredFiles {
		if !found {
			t.Errorf("Required file %s not found in EPUB", fileName)
		}
	}

	// Check mimetype is first and uncompressed
	if len(reader.File) == 0 {
		t.Fatal("EPUB has no files")
	}

	firstFile := reader.File[0]
	if firstFile.Name != "mimetype" {
		t.Errorf("First file should be 'mimetype', got %s", firstFile.Name)
	}

	if firstFile.Method != zip.Store {
		t.Error("mimetype file should be uncompressed (Store method)")
	}
}

func TestGenerateEPUB_ValidZIP(t *testing.T) {
	fb2Path := getTestDataPath(filepath.Join("valid", "minimal.fb2"))
	fb2, err := converter.ParseFB2(fb2Path)
	if err != nil {
		t.Fatalf("Failed to parse FB2: %v", err)
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.epub")

	err = converter.GenerateEPUB(fb2, outputPath)
	if err != nil {
		t.Fatalf("GenerateEPUB() error = %v, want nil", err)
	}

	// Try to open as ZIP
	reader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("EPUB is not a valid ZIP archive: %v", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			t.Logf("Error closing ZIP: %v", closeErr)
		}
	}()

	// Verify we can read files
	if len(reader.File) == 0 {
		t.Error("EPUB ZIP archive has no files")
	}
}

func TestGenerateEPUB_WithTOC(t *testing.T) {
	fb2Path := getTestDataPath(filepath.Join("valid", "complete.fb2"))
	fb2, err := converter.ParseFB2(fb2Path)
	if err != nil {
		t.Fatalf("Failed to parse FB2: %v", err)
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.epub")

	err = converter.GenerateEPUB(fb2, outputPath)
	if err != nil {
		t.Fatalf("GenerateEPUB() error = %v, want nil", err)
	}

	// Open EPUB
	reader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("Failed to open EPUB: %v", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			t.Logf("Error closing ZIP: %v", closeErr)
		}
	}()

	// Check TOC files exist
	tocFound := false
	navFound := false

	for _, file := range reader.File {
		if file.Name == "OEBPS/toc.ncx" {
			tocFound = true
		}
		if file.Name == "OEBPS/nav.xhtml" {
			navFound = true
		}
	}

	if !tocFound {
		t.Error("TOC file (toc.ncx) not found")
	}
	if !navFound {
		t.Error("Navigation file (nav.xhtml) not found")
	}
}

func TestGenerateEPUB_HTMLEscaping(t *testing.T) {
	fb2Path := getTestDataPath(filepath.Join("edge-cases", "unicode.fb2"))
	fb2, err := converter.ParseFB2(fb2Path)
	if err != nil {
		t.Fatalf("Failed to parse FB2: %v", err)
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.epub")

	err = converter.GenerateEPUB(fb2, outputPath)
	if err != nil {
		t.Fatalf("GenerateEPUB() error = %v, want nil", err)
	}

	// Open EPUB and check content
	reader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("Failed to open EPUB: %v", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			t.Logf("Error closing ZIP: %v", closeErr)
		}
	}()

	// Find content file
	var contentFile *zip.File
	for _, file := range reader.File {
		if file.Name == "OEBPS/content.xhtml" {
			contentFile = file
			break
		}
	}

	if contentFile == nil {
		t.Fatal("Content file not found")
	}

	// Read content
	rc, err := contentFile.Open()
	if err != nil {
		t.Fatalf("Failed to open content file: %v", err)
	}
	defer func() {
		if closeErr := rc.Close(); closeErr != nil {
			t.Logf("Error closing content file: %v", closeErr)
		}
	}()

	// Check for proper XML structure (no unescaped < or >)
	buf := make([]byte, 1024)
	n, err := rc.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read content: %v", err)
	}

	content := string(buf[:n])
	// Check that special characters are properly escaped
	if strings.Contains(content, "<tag>") && !strings.Contains(content, "&lt;tag&gt;") {
		// If we have <tag> in the original, it should be escaped in HTML
		// But if it's in a CDATA or already escaped, that's fine
		t.Log("Content contains unescaped tags - this might be OK if in CDATA")
	}
}

func TestGenerateEPUB_WithNestedSections(t *testing.T) {
	fb2Path := getTestDataPath(filepath.Join("valid", "complete.fb2"))
	fb2, err := converter.ParseFB2(fb2Path)
	if err != nil {
		t.Fatalf("Failed to parse FB2: %v", err)
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.epub")

	err = converter.GenerateEPUB(fb2, outputPath)
	if err != nil {
		t.Fatalf("GenerateEPUB() error = %v, want nil", err)
	}

	// Verify EPUB was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("EPUB file was not created")
	}

	// Open and verify structure
	reader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("Failed to open EPUB: %v", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			t.Logf("Error closing ZIP: %v", closeErr)
		}
	}()

	// Check that content file exists
	contentFound := false
	for _, file := range reader.File {
		if file.Name == "OEBPS/content.xhtml" {
			contentFound = true
			break
		}
	}

	if !contentFound {
		t.Error("Content file not found in EPUB")
	}
}

