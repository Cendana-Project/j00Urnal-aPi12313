package email

import (
	"strings"
	"testing"
)

func TestRenderVerifyPIN_Escaping(t *testing.T) {
	// Name contains HTML tag
	maliciousName := "<b>Hacker</b>"
	html := RenderVerifyPIN(maliciousName, "123456", 10)

	// Should NOT contain raw <b> tag
	if strings.Contains(html, "<b>") {
		t.Errorf("Expected HTML entity escaping, but found raw HTML tag: %s", html)
	}

	// Should contain escaped version
	if !strings.Contains(html, "&lt;b&gt;Hacker&lt;/b&gt;") {
		t.Errorf("Expected escaped name, but not found in: %s", html)
	}
}

func TestRenderResetPIN_Escaping(t *testing.T) {
	maliciousName := "<script>alert(1)</script>"
	html := RenderResetPIN(maliciousName, "654321", 5)

	if strings.Contains(html, "<script>") {
		t.Errorf("Expected HTML entity escaping, but found raw script tag: %s", html)
	}

	if !strings.Contains(html, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Errorf("Expected escaped script, but not found in: %s", html)
	}
}
