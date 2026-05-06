package email

import (
	"strings"
	"unicode/utf8"
)

// NormalizeSubject replaces common Unicode punctuation with ASCII and collapses whitespace
// so email subjects behave consistently across providers and RFC 2047 encoding.
func NormalizeSubject(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"\u2014", "-", // em dash
		"\u2013", "-", // en dash
		"\u00a0", " ", // nbsp
		"\u2026", "...", // ellipsis
		"\r\n", " ",
		"\n", " ",
		"\r", " ",
	)
	s = replacer.Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

// ReviewerInviteSubject builds a short, readable subject line. Long titles are truncated
// so clients are less likely to show awkward RFC2047 fragments in the inbox list.
func ReviewerInviteSubject(manuscriptTitle string) string {
	title := NormalizeSubject(manuscriptTitle)
	const brand = "MedikaOne Journal"
	if title == "" {
		return "Undangan review manuskrip - " + brand
	}
	const maxRunes = 55
	if utf8.RuneCountInString(title) > maxRunes {
		r := []rune(title)
		title = string(r[:maxRunes-1]) + "…"
	}
	return "Undangan review: " + title + " - " + brand
}

// ReviewerAssignmentNotifySubject is used when an existing reviewer user is assigned (portal login, no token).
func ReviewerAssignmentNotifySubject(manuscriptTitle string) string {
	title := NormalizeSubject(manuscriptTitle)
	const brand = "MedikaOne Journal"
	if title == "" {
		return "Penugasan review manuskrip - " + brand
	}
	const maxRunes = 55
	if utf8.RuneCountInString(title) > maxRunes {
		r := []rune(title)
		title = string(r[:maxRunes-1]) + "…"
	}
	return "Penugasan review: " + title + " - " + brand
}
