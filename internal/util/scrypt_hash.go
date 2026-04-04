package util

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/scrypt"
)

// HashPasswordScrypt produces the same format as auth.Service (base64(key):base64(salt)).
func HashPasswordScrypt(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := scrypt.Key([]byte(password), salt, 1<<15, 8, 1, 64)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key) + ":" + base64.StdEncoding.EncodeToString(salt), nil
}
