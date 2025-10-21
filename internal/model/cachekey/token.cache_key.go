package cachekey

import "fmt"

func NewRefreshTokenCacheKey(userID, tokenID string) string {
	return fmt.Sprintf("refresh-token:userId:%s:tokenId:%s", userID, tokenID)
}

func NewAccessTokenBlacklistCacheKey(userID, tokenID string) string {
	return fmt.Sprintf("access-token-blacklist:userId:%s:tokenId:%s", userID, tokenID)
}

func NewAutoLoginTokenCacheKey(token string) string {
	return fmt.Sprintf("auto-login-token:%s", token)
}
