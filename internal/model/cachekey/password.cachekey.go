package cachekey

import "fmt"

func NewForgotPasswordRateLimitKey(email string) string {
	return fmt.Sprintf("forgot_password_rate_limit:%s", email)
}

func NewForgotPasswordIPRateLimitKey(ip string) string {
	return fmt.Sprintf("forgot_password_ip_rate_limit:%s", ip)
}

func NewForgotPasswordCacheKey(token string) string {
	return fmt.Sprintf("forgot_password:%s", token)
}

func NewForgotPasswordOTPCacheKey(email string) string {
	return fmt.Sprintf("forgot_password_otp:%s", email)
}

func NewForgotPasswordVerifiedKey(email string) string {
	return fmt.Sprintf("forgot_password_verified:%s", email)
}

func NewForgotPasswordSessionKey(sessionToken string) string {
	return fmt.Sprintf("forgot_password_session:%s", sessionToken)
}

func NewPasswordHistoryCacheKey(userID string) string {
	return fmt.Sprintf("password_history:%s", userID)
}

func NewUserSessionCacheKey(userID string) string {
	return fmt.Sprintf("user_session:%s", userID)
}
