package util

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

// ValidateFile checks if the uploaded file meets the size, mime type, and extension requirements.
func ValidateFile(header *multipart.FileHeader, maxBytes int64, allowedMimes []string) error {
	// 1. Check Size
	if header.Size > maxBytes {
		mb := float64(maxBytes) / 1024 / 1024
		return fmt.Errorf("file size exceeds the maximum limit of %.2f MB", mb)
	}

	// 2. Check Mime Type (if restrictions apply)
	if len(allowedMimes) > 0 {
		src, err := header.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		// Read first 512 bytes for content type detection
		buffer := make([]byte, 512)
		_, err = src.Read(buffer)
		if err != nil && err.Error() != "EOF" { // EOF is fine for small files
			return err
		}

		// Detect Content-Type
		contentType := http.DetectContentType(buffer)

		// Check if detected type is allowed
		isAllowed := false
		for _, mime := range allowedMimes {
			// Handle simple prefix matching for broader types if needed,
			// but exact match is safer.
			// However, http.DetectContentType returns standard signatures.
			// Let's compare ignoring charset parameters usually found in text.
			baseType := strings.Split(contentType, ";")[0]
			if baseType == mime {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return fmt.Errorf("file type '%s' is not allowed. Allowed types: %v", contentType, allowedMimes)
		}

		// 3. Check Extension (to match content type somewhat and prevent weird renames)
		// We infer allowed extensions from allowedMimes or just strict check?
		// Since we didn't pass allowedExtensions, let's map known mimes to extensions or just rely on the fact
		// that we should reject .xxx if it doesn't match the mime?
		// User specifically complained about .xxx.
		// Simple fix: Ensure extension is valid for the detected mime?
		// Or simpler: Just check if the extension matches one of the common extensions for the allowed mimes.

		// Let's create a simple map for this project context
		ext := strings.ToLower(filepath.Ext(header.Filename))
		validExt := false

		// Common extensions map
		mimeToExt := map[string][]string{
			"image/jpeg":         {".jpg", ".jpeg"},
			"image/png":          {".png"},
			"image/webp":         {".webp"},
			"application/pdf":    {".pdf"},
			"application/msword": {".doc"},
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {".docx"},
			"text/csv":                 {".csv"},
			"text/plain":               {".txt"}, // Assuming txt for plain
			"application/vnd.ms-excel": {".xls"},
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": {".xlsx"},
		}

		for _, mime := range allowedMimes {
			if allowedExts, ok := mimeToExt[mime]; ok {
				for _, allowed := range allowedExts {
					if ext == allowed {
						validExt = true
						break
					}
				}
			}
		}
		// If allow list has something we assume we want to validate it.
		// However, if we missed a mapping, we might block valid files.
		// Safe bet: If the extension is .xxx, it definitely won't match common mappings.
		if !validExt && len(allowedMimes) > 0 {
			// Double check if we missed mapping?
			// For now, let's be strict as user requested.
			return fmt.Errorf("file extension '%s' is not valid for the allowed mime types", ext)
		}
	}

	return nil
}
