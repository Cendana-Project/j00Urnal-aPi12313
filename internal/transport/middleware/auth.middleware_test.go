package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
)

const testAccessTokenSecret = "test-access-token-secret"

type fakeBlacklist struct {
	result int64
	err    error
	keys   []string
}

func (f *fakeBlacklist) Exists(_ context.Context, keys ...string) *redis.IntCmd {
	f.keys = append(f.keys, keys...)
	return redis.NewIntResult(f.result, f.err)
}

func withTestTokenConfig(t *testing.T) {
	t.Helper()
	previous := config.Env
	config.Env = &config.EnvConfig{Token: config.Token{AccessTokenSecret: testAccessTokenSecret}}
	t.Cleanup(func() { config.Env = previous })
}

func signAccessToken(t *testing.T, method jwt.SigningMethod, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	signed, err := token.SignedString([]byte(testAccessTokenSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func validAccessClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"sub": "e1153298-e753-4e1b-b286-2f899a6499e8",
		"typ": "access",
		"jti": "b829588c-ad50-44ac-a5af-f50f6c0794eb",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
}

func exerciseMiddleware(
	t *testing.T,
	middleware gin.HandlerFunc,
	authorization string,
) (int, string, string, bool) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	reached := false
	userID := ""
	tokenID := ""
	router := gin.New()
	router.GET("/archive", middleware, func(c *gin.Context) {
		reached = true
		userID = c.GetString(string(constant.UserID))
		tokenID = c.GetString(string(constant.TokenID))
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/archive", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder.Code, userID, tokenID, reached
}

func TestAuthOptionalValidTokenSetsIdentityAndChecksBlacklist(t *testing.T) {
	withTestTokenConfig(t)
	claims := validAccessClaims()
	token := signAccessToken(t, jwt.SigningMethodHS256, claims)
	blacklist := &fakeBlacklist{}

	status, userID, tokenID, reached := exerciseMiddleware(
		t,
		authOptionalWithBlacklist(blacklist),
		"Bearer "+token,
	)

	if status != http.StatusNoContent || !reached {
		t.Fatalf("public handler was not reached: status=%d reached=%v", status, reached)
	}
	if userID != claims["sub"] || tokenID != claims["jti"] {
		t.Fatalf("unexpected identity: user_id=%q jti=%q", userID, tokenID)
	}
	wantKey := accessBlacklistKey(claims["jti"].(string))
	if len(blacklist.keys) != 1 || blacklist.keys[0] != wantKey {
		t.Fatalf("blacklist keys = %v, want [%q]", blacklist.keys, wantKey)
	}
}

func TestAuthOptionalInvalidCredentialsRemainAnonymous(t *testing.T) {
	withTestTokenConfig(t)

	expiredClaims := validAccessClaims()
	expiredClaims["exp"] = time.Now().Add(-time.Hour).Unix()
	wrongTypeClaims := validAccessClaims()
	wrongTypeClaims["typ"] = "refresh"
	missingSubjectClaims := validAccessClaims()
	delete(missingSubjectClaims, "sub")
	missingJTIClaims := validAccessClaims()
	delete(missingJTIClaims, "jti")
	missingExpiryClaims := validAccessClaims()
	delete(missingExpiryClaims, "exp")

	tests := []struct {
		name          string
		authorization string
	}{
		{name: "missing header"},
		{name: "wrong scheme", authorization: "Basic credentials"},
		{name: "malformed token", authorization: "Bearer not-a-jwt"},
		{name: "expired token", authorization: "Bearer " + signAccessToken(t, jwt.SigningMethodHS256, expiredClaims)},
		{name: "refresh token", authorization: "Bearer " + signAccessToken(t, jwt.SigningMethodHS256, wrongTypeClaims)},
		{name: "missing subject", authorization: "Bearer " + signAccessToken(t, jwt.SigningMethodHS256, missingSubjectClaims)},
		{name: "missing jti", authorization: "Bearer " + signAccessToken(t, jwt.SigningMethodHS256, missingJTIClaims)},
		{name: "missing expiry", authorization: "Bearer " + signAccessToken(t, jwt.SigningMethodHS256, missingExpiryClaims)},
		{name: "wrong signing algorithm", authorization: "Bearer " + signAccessToken(t, jwt.SigningMethodHS384, validAccessClaims())},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, userID, tokenID, reached := exerciseMiddleware(t, AuthOptional(nil), tt.authorization)
			if status != http.StatusNoContent || !reached {
				t.Fatalf("public handler was not reached: status=%d reached=%v", status, reached)
			}
			if userID != "" || tokenID != "" {
				t.Fatalf("invalid credentials authenticated as user_id=%q jti=%q", userID, tokenID)
			}
		})
	}
}

func TestAuthOptionalBlacklistFailureCannotGrantPrivilegedIdentity(t *testing.T) {
	withTestTokenConfig(t)
	token := signAccessToken(t, jwt.SigningMethodHS256, validAccessClaims())

	tests := []struct {
		name      string
		blacklist *fakeBlacklist
	}{
		{name: "revoked", blacklist: &fakeBlacklist{result: 1}},
		{name: "redis unavailable", blacklist: &fakeBlacklist{err: errors.New("redis unavailable")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, userID, tokenID, reached := exerciseMiddleware(
				t,
				authOptionalWithBlacklist(tt.blacklist),
				"Bearer "+token,
			)
			if status != http.StatusNoContent || !reached {
				t.Fatalf("public handler was not reached: status=%d reached=%v", status, reached)
			}
			if userID != "" || tokenID != "" {
				t.Fatalf("unsafe identity on blacklist failure: user_id=%q jti=%q", userID, tokenID)
			}
		})
	}
}

func TestAuthOptionalValidTokenWithDisabledRedis(t *testing.T) {
	withTestTokenConfig(t)
	claims := validAccessClaims()
	token := signAccessToken(t, jwt.SigningMethodHS256, claims)

	status, userID, tokenID, reached := exerciseMiddleware(t, AuthOptional(nil), "Bearer "+token)
	if status != http.StatusNoContent || !reached {
		t.Fatalf("public handler was not reached: status=%d reached=%v", status, reached)
	}
	if userID != claims["sub"] || tokenID != claims["jti"] {
		t.Fatalf("unexpected identity: user_id=%q jti=%q", userID, tokenID)
	}
}

func TestAuthRequiredValidTokenWithDisabledRedis(t *testing.T) {
	withTestTokenConfig(t)
	claims := validAccessClaims()
	token := signAccessToken(t, jwt.SigningMethodHS256, claims)

	status, userID, tokenID, reached := exerciseMiddleware(t, AuthRequired(nil), "Bearer "+token)
	if status != http.StatusNoContent || !reached {
		t.Fatalf("protected handler was not reached: status=%d reached=%v", status, reached)
	}
	if userID != claims["sub"] || tokenID != claims["jti"] {
		t.Fatalf("unexpected identity: user_id=%q jti=%q", userID, tokenID)
	}
}

func TestAuthRequiredStillRejectsMissingAndBlacklistedTokens(t *testing.T) {
	withTestTokenConfig(t)
	token := signAccessToken(t, jwt.SigningMethodHS256, validAccessClaims())

	tests := []struct {
		name          string
		authorization string
		blacklist     *fakeBlacklist
	}{
		{name: "missing"},
		{name: "blacklisted", authorization: "Bearer " + token, blacklist: &fakeBlacklist{result: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _, _, reached := exerciseMiddleware(t, authRequiredWithBlacklist(tt.blacklist), tt.authorization)
			if status != constant.ErrUnauthorized.StatusCode {
				t.Fatalf("status = %d, want %d", status, constant.ErrUnauthorized.StatusCode)
			}
			if reached {
				t.Fatal("protected handler was reached")
			}
		})
	}
}
