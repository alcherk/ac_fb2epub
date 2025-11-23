package converter

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"github.com/lex/fb2epub/models"
)

// ParseFB2 parses an FB2 file and returns a FictionBook struct
func ParseFB2(filePath string) (*models.FictionBook, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	var fb2 models.FictionBook
	decoder := xml.NewDecoder(file)

	// Handle XML namespaces and encoding
	decoder.CharsetReader = func(_ string, input io.Reader) (io.Reader, error) {
		return input, nil
	}

	if err := decoder.Decode(&fb2); err != nil {
		return nil, fmt.Errorf("failed to parse FB2 XML: %w", err)
	}

	return &fb2, nil
}

// ParseFB2FromReader parses FB2 from an io.Reader
func ParseFB2FromReader(reader io.Reader) (*models.FictionBook, error) {
	var fb2 models.FictionBook
	decoder := xml.NewDecoder(reader)

	decoder.CharsetReader = func(_ string, input io.Reader) (io.Reader, error) {
		return input, nil
	}

	if err := decoder.Decode(&fb2); err != nil {
		return nil, fmt.Errorf("failed to parse FB2 XML: %w", err)
	}

	return &fb2, nil
}
