package converter_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/lex/fb2epub/converter"
)

func getTestDataPath(filename string) string {
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	projectRoot := filepath.Join(testDir, "..", "..")
	return filepath.Join(projectRoot, "testdata", filename)
}

func TestEPUBContent_HTMLEscaping(t *testing.T) {
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

	buf := make([]byte, 4096)
	n, err := rc.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read content: %v", err)
	}

	content := string(buf[:n])

	// Check that HTML entities are properly escaped
	if strings.Contains(content, "&lt;tag&gt;") {
		// Good - properly escaped
	} else if strings.Contains(content, "<tag>") && !strings.Contains(content, "<![CDATA[") {
		// If we have unescaped tags outside CDATA, that's a problem
		t.Error("HTML special characters should be properly escaped")
	}
}

func TestEPUBContent_SectionIDs(t *testing.T) {
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

	// Check TOC file for section IDs
	var tocFile *zip.File
	for _, file := range reader.File {
		if file.Name == "OEBPS/toc.ncx" {
			tocFile = file
			break
		}
	}

	if tocFile == nil {
		t.Fatal("TOC file not found")
	}

	rc, err := tocFile.Open()
	if err != nil {
		t.Fatalf("Failed to open TOC file: %v", err)
	}
	defer func() {
		if closeErr := rc.Close(); closeErr != nil {
			t.Logf("Error closing TOC file: %v", closeErr)
		}
	}()

	buf := make([]byte, 2048)
	n, err := rc.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read TOC: %v", err)
	}

	tocContent := string(buf[:n])

	// Check that section IDs are present and properly formatted
	if !strings.Contains(tocContent, "id=\"") {
		t.Error("TOC should contain section IDs")
	}
}

func TestEPUBContent_ImageReferences(t *testing.T) {
	// This test would require an FB2 file with images
	// For now, we'll test that the EPUB structure supports images
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

	// Open EPUB and check manifest
	reader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("Failed to open EPUB: %v", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			t.Logf("Error closing ZIP: %v", closeErr)
		}
	}()

	// Check content.opf for image manifest items
	var opfFile *zip.File
	for _, file := range reader.File {
		if file.Name == "OEBPS/content.opf" {
			opfFile = file
			break
		}
	}

	if opfFile == nil {
		t.Fatal("OPF file not found")
	}

	rc, err := opfFile.Open()
	if err != nil {
		t.Fatalf("Failed to open OPF file: %v", err)
	}
	defer func() {
		if closeErr := rc.Close(); closeErr != nil {
			t.Logf("Error closing OPF file: %v", closeErr)
		}
	}()

	buf := make([]byte, 4096)
	n, err := rc.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read OPF: %v", err)
	}

	opfContent := string(buf[:n])

	// Check that manifest structure is present
	if !strings.Contains(opfContent, "<manifest>") {
		t.Error("OPF should contain manifest section")
	}

	if !strings.Contains(opfContent, "<item") {
		t.Error("OPF manifest should contain items")
	}
}

func TestEPUBContent_LinksFormatting(t *testing.T) {
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

	// Open EPUB and check content for links
	reader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("Failed to open EPUB: %v", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			t.Logf("Error closing ZIP: %v", closeErr)
		}
	}()

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

	rc, err := contentFile.Open()
	if err != nil {
		t.Fatalf("Failed to open content file: %v", err)
	}
	defer func() {
		if closeErr := rc.Close(); closeErr != nil {
			t.Logf("Error closing content file: %v", closeErr)
		}
	}()

	buf := make([]byte, 4096)
	n, err := rc.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read content: %v", err)
	}

	content := string(buf[:n])

	// Check for link formatting (if links exist in the FB2)
	if strings.Contains(content, "http://example.com") {
		// If we have links, they should be properly formatted
		if !strings.Contains(content, "<a ") && !strings.Contains(content, "href=") {
			t.Error("Links should be properly formatted in HTML")
		}
	}
}

func TestEPUBContent_SpecialCharacters(t *testing.T) {
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

	// Verify EPUB was created successfully
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("EPUB file was not created")
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

	// Check that content file exists
	contentFound := false
	for _, file := range reader.File {
		if file.Name == "OEBPS/content.xhtml" {
			contentFound = true
			break
		}
	}

	if !contentFound {
		t.Error("Content file should exist in EPUB")
	}
}
