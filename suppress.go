package otelmetriclint

import (
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bit-mover/otelmetriclint/rules"
)

const linterName = "otelmetriclint"

// matchesNoLint reports whether commentText (raw text including the
// leading // or /* and trailing */) is a //nolint directive that
// suppresses this analyzer.
//
// Recognized grammar mirrors golangci-lint:
//
//	//nolint                          → suppress everything
//	//nolint:<linter>[,<linter>...]   → suppress listed linters
//	//nolint:<list> <reason text>     → reason text after the first "//"
//	                                    or "/*" in the list segment is
//	                                    ignored (no leading space required)
//
// Parser quirks (also from golangci):
//   - Leading "//" or "/*" required; no space between "//" and "nolint".
//   - Keyword is "nolint", case-sensitive.
//   - Linter names match exactly (no prefix/substring).
//   - "//nolint:" with an empty list matches nothing — it is NOT treated
//     as bare "//nolint". This mirrors current golangci-lint behavior.
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
	// Trim trailing reason text. Anything after the first "//" or "/*" in
	// the list segment is treated as reason text (no leading space required),
	// matching golangci-lint.
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

// suppressIndex answers "is this MetricCall suppressed?" in O(1) after
// construction. Built once per analyzer Run.
type suppressIndex struct {
	fset *token.FileSet
	// directiveEndLine[file] is the set of line numbers where a matching
	// //nolint directive ends. A trailing directive's start and end are
	// the same line; a multi-line block directive's end is the last line.
	directiveEndLine map[*token.File]map[int]bool
}

// buildSuppressIndex scans every comment in pass.Files for //nolint
// directives that suppress this analyzer and records each directive's
// end line.
func buildSuppressIndex(pass *analysis.Pass) suppressIndex {
	idx := suppressIndex{
		fset:             pass.Fset,
		directiveEndLine: make(map[*token.File]map[int]bool),
	}
	for _, file := range pass.Files {
		tokFile := pass.Fset.File(file.Pos())
		if tokFile == nil {
			continue
		}
		for _, group := range file.Comments {
			for _, c := range group.List {
				if !matchesNoLint(c.Text) {
					continue
				}
				endLine := pass.Fset.Position(c.End()).Line
				if idx.directiveEndLine[tokFile] == nil {
					idx.directiveEndLine[tokFile] = make(map[int]bool)
				}
				idx.directiveEndLine[tokFile][endLine] = true
			}
		}
	}
	return idx
}

// suppressed reports whether a MetricCall at call.Pos is covered by any
// recognized //nolint directive.
func (s suppressIndex) suppressed(call rules.MetricCall) bool {
	if !call.Pos.IsValid() {
		return false
	}
	tokFile := s.fset.File(call.Pos)
	if tokFile == nil {
		return false
	}
	lines := s.directiveEndLine[tokFile]
	if lines == nil {
		return false
	}
	callLine := s.fset.Position(call.Pos).Line
	// Trailing: directive on the same line as the call.
	if lines[callLine] {
		return true
	}
	// Above call: directive ends on the line immediately preceding the call.
	// A blank line between directive and call breaks adjacency (golangci's rule).
	// Further placement branches (above-func, above-package) follow.
	if lines[callLine-1] {
		return true
	}
	return false
}
