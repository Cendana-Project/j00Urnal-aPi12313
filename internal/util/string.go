package util

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/api-monolith-template/internal/constant"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/scrypt"
)

type TokenClaims struct {
	UserID string             `json:"user_id"`
	Type   constant.TokenType `json:"type"`
	jwt.RegisteredClaims
}

func HashPassword(password string, salt []byte) (string, error) {
	hash, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", err
	}
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("%s.%s", encodedSalt, encodedHash), nil
}

func ComparePassword(hashedPassword, password string) (bool, error) {
	parts := strings.Split(hashedPassword, ".")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid hashed password format")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	newHash, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare(hash, newHash) == 1, nil
}

func GenerateToken(secret, userID, tokenID string, duration time.Duration, tokenType constant.TokenType) (string, time.Time, error) {
	expiredAt := time.Now().Add(duration)

	claims := TokenClaims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(expiredAt),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	return tokenString, expiredAt, err
}

func GenerateRandomToken(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

func GenerateOTP(length int) string {
	const charset = "0123456789"
	result := make([]byte, length)

	for i := range result {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			index = big.NewInt(time.Now().UnixNano() % int64(len(charset)))
		}
		result[i] = charset[index.Int64()]
	}

	return string(result)
}
