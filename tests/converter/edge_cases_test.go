package converter_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lex/fb2epub/converter"
)

func TestEdgeCase_NoSections(t *testing.T) {
	// Create a test FB2 file with no sections
	fb2Content := `<?xml version="1.0" encoding="UTF-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
  <description>
    <title-info>
      <book-title>Book Without Sections</book-title>
      <author>
        <first-name>Test</first-name>
        <last-name>Author</last-name>
      </author>
    </title-info>
  </description>
  <body>
  </body>
</FictionBook>`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "no-sections.fb2")
	if err := os.WriteFile(testFile, []byte(fb2Content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fb2, err := converter.ParseFB2(testFile)
	if err != nil {
		t.Fatalf("ParseFB2() error = %v, want nil", err)
	}

	if fb2 == nil {
		t.Fatal("ParseFB2() returned nil")
	}

	// EPUB generation should still work even with no sections
	outputPath := filepath.Join(tmpDir, "output.epub")
	err = converter.GenerateEPUB(fb2, outputPath)
	if err != nil {
		t.Fatalf("GenerateEPUB() should handle no sections: %v", err)
	}
}

func TestEdgeCase_NoTitle(t *testing.T) {
	// Create a test FB2 file with no title
	fb2Content := `<?xml version="1.0" encoding="UTF-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
  <description>
    <title-info>
      <author>
        <first-name>Test</first-name>
        <last-name>Author</last-name>
      </author>
    </title-info>
  </description>
  <body>
    <section>
      <p>Content without title.</p>
    </section>
  </body>
</FictionBook>`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "no-title.fb2")
	if err := os.WriteFile(testFile, []byte(fb2Content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fb2, err := converter.ParseFB2(testFile)
	if err != nil {
		t.Fatalf("ParseFB2() error = %v, want nil", err)
	}

	if fb2 == nil {
		t.Fatal("ParseFB2() returned nil")
	}

	// EPUB generation should use default title
	outputPath := filepath.Join(tmpDir, "output.epub")
	err = converter.GenerateEPUB(fb2, outputPath)
	if err != nil {
		t.Fatalf("GenerateEPUB() should handle no title: %v", err)
	}
}

func TestEdgeCase_LongText(t *testing.T) {
	// Create a test FB2 file with very long text
	longText := strings.Repeat("This is a very long paragraph. ", 1000)
	
	fb2Content := `<?xml version="1.0" encoding="UTF-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
  <description>
    <title-info>
      <book-title>Book with Long Text</book-title>
      <author>
        <first-name>Test</first-name>
        <last-name>Author</last-name>
      </author>
    </title-info>
  </description>
  <body>
    <section>
      <p>` + longText + `</p>
    </section>
  </body>
</FictionBook>`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "long-text.fb2")
	if err := os.WriteFile(testFile, []byte(fb2Content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fb2, err := converter.ParseFB2(testFile)
	if err != nil {
		t.Fatalf("ParseFB2() error = %v, want nil", err)
	}

	outputPath := filepath.Join(tmpDir, "output.epub")
	err = converter.GenerateEPUB(fb2, outputPath)
	if err != nil {
		t.Fatalf("GenerateEPUB() should handle long text: %v", err)
	}

	// Verify EPUB was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("EPUB file was not created")
	}
}

func TestEdgeCase_Emojis(t *testing.T) {
	// Create a test FB2 file with emoji characters
	fb2Content := `<?xml version="1.0" encoding="UTF-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
  <description>
    <title-info>
      <book-title>Book with Emojis 😀📚</book-title>
      <author>
        <first-name>Test</first-name>
        <last-name>Author</last-name>
      </author>
    </title-info>
  </description>
  <body>
    <section>
      <p>Text with emojis: 😀 🎉 📚 🚀 ⭐</p>
      <p>More emojis: 🎨 🎭 🎪 🎯 🎲</p>
    </section>
  </body>
</FictionBook>`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "emojis.fb2")
	if err := os.WriteFile(testFile, []byte(fb2Content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fb2, err := converter.ParseFB2(testFile)
	if err != nil {
		t.Fatalf("ParseFB2() error = %v, want nil", err)
	}

	if fb2 == nil {
		t.Fatal("ParseFB2() returned nil")
	}

	outputPath := filepath.Join(tmpDir, "output.epub")
	err = converter.GenerateEPUB(fb2, outputPath)
	if err != nil {
		t.Fatalf("GenerateEPUB() should handle emojis: %v", err)
	}

	// Verify EPUB was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("EPUB file was not created")
	}
}

func TestEdgeCase_DeepNesting(t *testing.T) {
	// Create a test FB2 file with deeply nested sections
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
  <description>
    <title-info>
      <book-title>Book with Deep Nesting</book-title>
      <author>
        <first-name>Test</first-name>
        <last-name>Author</last-name>
      </author>
    </title-info>
  </description>
  <body>
    <section>
      <title>
        <p>Root Section</p>
      </title>
      <p>Root content</p>
`)
	
	// Build nested sections
	sectionDepth := 5
	for i := 1; i <= sectionDepth; i++ {
		indent := strings.Repeat("      ", i)
		builder.WriteString(indent + `<section>
`)
		builder.WriteString(indent + `  <title>
`)
		builder.WriteString(indent + `    <p>Section Level `)
		builder.WriteString(fmt.Sprintf("%d", i))
		builder.WriteString(`</p>
`)
		builder.WriteString(indent + `  </title>
`)
		builder.WriteString(indent + `  <p>Content at level `)
		builder.WriteString(fmt.Sprintf("%d", i))
		builder.WriteString(`</p>
`)
	}
	
	// Close nested sections
	for i := sectionDepth; i >= 1; i-- {
		indent := strings.Repeat("      ", i)
		builder.WriteString(indent + `</section>
`)
	}
	
	builder.WriteString(`    </section>
  </body>
</FictionBook>`)
	
	fb2Content := builder.String()
	
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "deep-nesting.fb2")
	if err := os.WriteFile(testFile, []byte(fb2Content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fb2, err := converter.ParseFB2(testFile)
	if err != nil {
		t.Fatalf("ParseFB2() error = %v, want nil", err)
	}

	if fb2 == nil {
		t.Fatal("ParseFB2() returned nil")
	}

	outputPath := filepath.Join(tmpDir, "output.epub")
	err = converter.GenerateEPUB(fb2, outputPath)
	if err != nil {
		t.Fatalf("GenerateEPUB() should handle deep nesting: %v", err)
	}

	// Verify EPUB was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("EPUB file was not created")
	}
}

