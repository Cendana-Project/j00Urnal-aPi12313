package cachekey

import (
	"fmt"

	"github.com/api-monolith-template/internal/constant"
)

func NewUserByIdentifierCacheKey(identifier string) string {
	return fmt.Sprintf("user:identifier:%s", identifier)
}

func NewUserByIDCacheKey(id string) string {
	return fmt.Sprintf("user:id:%s", id)
}

func NewEmailVerificationOTPCacheKey(email string) string {
	return fmt.Sprintf("email_verification_otp:%s", email)
}

func NewUserNonPrimaryKeyCacheKeysPatterns() []string {
	return []string{
		NewUserByIdentifierCacheKey("*"),
		constant.UserAllCacheKey,
	}
}
