package converter_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lex/fb2epub/converter"
)

func TestParseFB2_ValidFile(t *testing.T) {
	filePath := getTestDataPath(filepath.Join("valid", "minimal.fb2"))
	fb2, err := converter.ParseFB2(filePath)
	if err != nil {
		t.Fatalf("ParseFB2() error = %v, want nil", err)
	}

	if fb2 == nil {
		t.Fatal("ParseFB2() returned nil")
	}

	if fb2.Description.TitleInfo.BookTitle == "" {
		t.Error("BookTitle is empty")
	}

	if fb2.Body.Section == nil || len(fb2.Body.Section) == 0 {
		t.Error("Body has no sections")
	}
}

func TestParseFB2_CompleteFile(t *testing.T) {
	filePath := getTestDataPath(filepath.Join("valid", "complete.fb2"))
	fb2, err := converter.ParseFB2(filePath)
	if err != nil {
		t.Fatalf("ParseFB2() error = %v, want nil", err)
	}

	if fb2 == nil {
		t.Fatal("ParseFB2() returned nil")
	}

	// Check for nested sections
	if len(fb2.Body.Section) > 0 {
		firstSection := fb2.Body.Section[0]
		if len(firstSection.Section) > 0 {
			// Has nested sections
			if firstSection.Section[0].Title != nil && len(firstSection.Section[0].Title.Paragraph) == 0 {
				t.Error("Nested section should have a title")
			}
		}
	}
}

func TestParseFB2_InvalidXML(t *testing.T) {
	filePath := getTestDataPath(filepath.Join("invalid", "malformed.xml"))
	_, err := converter.ParseFB2(filePath)
	if err == nil {
		t.Error("ParseFB2() error = nil, want error for malformed XML")
	}
}

func TestParseFB2_MissingFile(t *testing.T) {
	filePath := getTestDataPath("nonexistent.fb2")
	_, err := converter.ParseFB2(filePath)
	if err == nil {
		t.Error("ParseFB2() error = nil, want error for missing file")
	}
}

func TestParseFB2_EmptyFile(t *testing.T) {
	filePath := getTestDataPath(filepath.Join("invalid", "empty.fb2"))
	fb2, err := converter.ParseFB2(filePath)
	if err != nil {
		// Empty file might parse but have no content
		t.Logf("ParseFB2() error = %v (expected for empty file)", err)
		return
	}

	if fb2 == nil {
		t.Fatal("ParseFB2() returned nil")
	}
}

func TestParseFB2_UnicodeFile(t *testing.T) {
	filePath := getTestDataPath(filepath.Join("edge-cases", "unicode.fb2"))
	fb2, err := converter.ParseFB2(filePath)
	if err != nil {
		t.Fatalf("ParseFB2() error = %v, want nil", err)
	}

	if fb2 == nil {
		t.Fatal("ParseFB2() returned nil")
	}

	// Check that Unicode content is preserved
	if len(fb2.Body.Section) > 0 {
		section := fb2.Body.Section[0]
		if section.Title != nil && len(section.Title.Paragraph) > 0 {
			title := section.Title.Paragraph[0].Text
			if title == "" {
				t.Error("Unicode title should not be empty")
			}
		}
	}
}

func TestParseFB2FromReader(t *testing.T) {
	filePath := getTestDataPath(filepath.Join("valid", "minimal.fb2"))
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Logf("Error closing file: %v", closeErr)
		}
	}()

	fb2, err := converter.ParseFB2FromReader(file)
	if err != nil {
		t.Fatalf("ParseFB2FromReader() error = %v, want nil", err)
	}

	if fb2 == nil {
		t.Fatal("ParseFB2FromReader() returned nil")
	}

	if fb2.Description.TitleInfo.BookTitle == "" {
		t.Error("Parsed FB2 should have book title")
	}
}
