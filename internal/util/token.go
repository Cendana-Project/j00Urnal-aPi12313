package util

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateSecureToken generates a URL-safe random token of given byte length.
// Reused by auth (PIN), review (invitation tokens), etc.
func GenerateSecureToken(byteLength int) string {
	b := make([]byte, byteLength)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
