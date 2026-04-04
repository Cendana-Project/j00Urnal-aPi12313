package util

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

// GenerateSecureToken generates a URL-safe random token of given byte length.
// Uses RawURLEncoding (no '=' padding) so tokens survive path/query handling in HTTP clients.
// Reused by review (invitation tokens), etc.
func GenerateSecureToken(byteLength int) string {
	b := make([]byte, byteLength)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// InvitationTokenCandidates returns strings to match against DB invitation_token.
// Standard base64url adds trailing '=' padding; path params and some clients strip it, so lookups must try variants.
func InvitationTokenCandidates(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	seen := make(map[string]struct{})
	var out []string
	add := func(s string) {
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	add(raw)
	if m := len(raw) % 4; m != 0 {
		add(raw + strings.Repeat("=", 4-m))
	}
	stripped := strings.TrimRight(raw, "=")
	if stripped != raw {
		add(stripped)
		if m := len(stripped) % 4; m != 0 {
			add(stripped + strings.Repeat("=", 4-m))
		}
	}
	return out
}
