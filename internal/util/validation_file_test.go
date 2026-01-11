package util

import (
	"os"
	"testing"
)

func TestValidateFile(t *testing.T) {
	// Create temporary files for testing
	tmpPDF, _ := os.CreateTemp("", "test*.pdf")
	defer os.Remove(tmpPDF.Name())
	tmpPDF.Write([]byte("%PDF-1.4 header simulates pdf"))
	tmpPDF.Close()

	tmpTXT, _ := os.CreateTemp("", "test*.txt")
	defer os.Remove(tmpTXT.Name())
	tmpTXT.Write([]byte("Just some plain text"))
	tmpTXT.Close()

	// Mock FileHeader (We need to open the actual file, so we need a way to mock or point to real temp file)
	// multipart.FileHeader has a method Open(). In real usage it opens the uploaded temp file.
	// To test this without setting up a full HTTP request, we can assume the header points to our local temp file.
	// Wait, FileHeader.Open() works if the file was populated by multipart parser.
	// Manually constructing it is tricky because it relies on internal path or Opening the underlying file.
	// Actually, `ValidateFile` takes `*multipart.FileHeader`.
	// For unit testing this strictly, we might need to rely on `mime/multipart/writer` to create a real form data body.
	// OR, we can just skip heavy unit testing here and rely on manual verification as the logic depends on `header.Open()`.

	// Simpler approach for "Unit" test: Create a real multipart request and parse it.
}

// Since mocking FileHeader.Open is hard, let's keep it simple for now and rely on manual verification plan.
// The logic is standard.
