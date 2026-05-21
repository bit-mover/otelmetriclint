package otelmetriclint

import "strings"

const linterName = "otelmetriclint"

// matchesNoLint reports whether commentText (raw text including the
// leading // or /* and trailing */) is a //nolint directive that
// suppresses this analyzer.
//
// Recognized grammar mirrors golangci-lint:
//
//	//nolint                          → suppress everything
//	//nolint:<linter>[,<linter>...]   → suppress listed linters
//	//nolint:<list> // reason text    → trailing reason is ignored
//
// Parser quirks (also from golangci):
//   - Leading "//" or "/*" required; no space between "//" and "nolint".
//   - Keyword is "nolint", case-sensitive.
//   - Linter names match exactly (no prefix/substring).
func matchesNoLint(commentText string) bool {
	body, ok := stripCommentDelimiters(commentText)
	if !ok {
		return false
	}
	if !strings.HasPrefix(body, "nolint") {
		return false
	}
	rest := body[len("nolint"):]
	// Bare //nolint (or //nolint followed by whitespace) suppresses everything.
	if rest == "" || isSpace(rest[0]) {
		return true
	}
	if rest[0] != ':' {
		return false // e.g. "nolintfoo"
	}
	listAndReason := rest[1:]
	// Trim a trailing " // reason" (or " /* reason */") — anything after
	// the first " //" or " /*" is reason text per golangci's convention.
	if idx := strings.Index(listAndReason, "//"); idx != -1 {
		listAndReason = listAndReason[:idx]
	}
	if idx := strings.Index(listAndReason, "/*"); idx != -1 {
		listAndReason = listAndReason[:idx]
	}
	for _, name := range strings.Split(listAndReason, ",") {
		if strings.TrimSpace(name) == linterName {
			return true
		}
	}
	return false
}

// stripCommentDelimiters removes the leading "//" or "/*" and the trailing
// "*/" (if a block comment), returning (body, true). Returns ("", false)
// if commentText doesn't begin with a recognized delimiter. A space between
// "//" and the next character disqualifies the comment (golangci's quirk).
func stripCommentDelimiters(commentText string) (string, bool) {
	switch {
	case strings.HasPrefix(commentText, "//"):
		body := commentText[2:]
		// The space-after-// quirk: "// nolint" doesn't count.
		if body == "" || isSpace(body[0]) {
			return "", false
		}
		return body, true
	case strings.HasPrefix(commentText, "/*"):
		body := strings.TrimSuffix(commentText[2:], "*/")
		body = strings.TrimSpace(body)
		return body, true
	}
	return "", false
}

func isSpace(b byte) bool { return b == ' ' || b == '\t' }
