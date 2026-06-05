package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strconv"
	"strings"
)

// UCUMUnit fires when a metric.WithUnit(...) call is given a string that is
// not a valid UCUM case-sensitive code. It applies to every instrument kind.
//
// The Allow field is an optional allowlist of non-UCUM strings that should be
// accepted silently (e.g. project-local conventions). Entries are matched
// against the unquoted literal value.
type UCUMUnit struct {
	Allow []string
}

// ID returns the rule's stable identifier.
func (UCUMUnit) ID() string { return "ucum_unit" }

// Check scans the call's options for WithUnit calls. If found, the string
// argument is validated against the UCUM spec. Diagnostics are emitted at the
// position of the string literal.
func (u UCUMUnit) Check(call MetricCall) []Diagnostic {
	var diags []Diagnostic
	for _, opt := range call.Options {
		if !isWithUnitCall(opt) {
			continue
		}
		callExpr, ok := opt.(*ast.CallExpr)
		if !ok {
			continue
		}
		if len(callExpr.Args) != 1 {
			continue
		}
		lit, ok := callExpr.Args[0].(*ast.BasicLit)
		if !ok {
			// non-literal argument: skip silently
			continue
		}
		if lit.Kind != token.STRING {
			continue
		}
		unquoted, err := strconv.Unquote(lit.Value)
		if err != nil {
			continue
		}
		// Check allowlist first.
		for _, a := range u.Allow {
			if unquoted == a {
				goto next
			}
		}
		if !validUCUM(unquoted) {
			suggestion := suggestUCUM(unquoted)
			var msg string
			if suggestion != "" {
				msg = fmt.Sprintf(
					"metric unit %q is not a valid UCUM code — did you mean %q?",
					unquoted, suggestion,
				)
			} else {
				msg = fmt.Sprintf(
					"metric unit %q is not a valid UCUM code — see https://ucum.org/ucum for valid codes",
					unquoted,
				)
			}
			diags = append(diags, Diagnostic{
				Pos:     lit.Pos(),
				Message: msg,
				RuleID:  u.ID(),
			})
		}
	next:
	}
	return diags
}

// validUCUM reports whether s is a syntactically valid UCUM case-sensitive
// code according to the simplified validation algorithm:
//
//  1. Empty string → valid.
//  2. Whole-string annotation {…} → valid.
//  3. Split on '.' and '/' into terms; each term must be valid (see below).
//
// A term is valid if:
//   - After stripping a trailing {annotation}, and
//   - After stripping a trailing optional-signed integer exponent,
//   - The remainder matches a ucumAtom directly, OR
//   - Stripping exactly one known prefix (longest-match first so binary
//     Ki/Mi beat single-char alternatives) and rechecking the remainder
//     against ucumAtoms succeeds.
func validUCUM(s string) bool {
	if s == "" {
		return true
	}
	// Whole-string annotation: {…}
	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		return true
	}
	// Split on '.' and '/' into terms.
	terms := splitUCUMTerms(s)
	for _, term := range terms {
		if !validUCUMTerm(term) {
			return false
		}
	}
	return true
}

// splitUCUMTerms splits s on '.' and '/' delimiters into individual terms.
// Annotations ({…}) are kept intact — the splitter does not split inside them.
func splitUCUMTerms(s string) []string {
	var terms []string
	depth := 0 // annotation nesting depth
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			if depth > 0 {
				depth--
			}
		case '.', '/':
			if depth == 0 {
				terms = append(terms, s[start:i])
				start = i + 1
			}
		}
	}
	terms = append(terms, s[start:])
	return terms
}

// validUCUMTerm validates a single UCUM term (no '.' or '/' delimiters).
func validUCUMTerm(term string) bool {
	if term == "" {
		return false
	}
	// Strip trailing annotation {…} first, then check.
	stripped := stripTrailingAnnotation(term)
	if stripped == "" {
		// The term was a pure {annotation} (e.g. "{packet}" in
		// "{packet}/s") — a dimensionless annotated unit, valid in UCUM.
		return true
	}
	// Strip trailing optional-signed integer exponent, but only if that
	// leaves a non-empty atom candidate. This prevents "1" (unity atom) from
	// being consumed as a bare exponent.
	candidate := stripTrailingExponent(stripped)
	if candidate == "" {
		// Exponent consumed the whole thing — treat the annotation-stripped
		// form as the atom candidate (e.g. "1" is a valid UCUM unity atom).
		candidate = stripped
	}
	// Direct atom lookup.
	if ucumAtoms[candidate] {
		return true
	}
	// Try stripping exactly one prefix (longest match first).
	prefixes := sortedPrefixesLongestFirst()
	for _, pfx := range prefixes {
		if strings.HasPrefix(candidate, pfx) {
			remainder := candidate[len(pfx):]
			if remainder != "" && ucumAtoms[remainder] {
				return true
			}
		}
	}
	return false
}

// stripTrailingAnnotation removes a trailing {annotation} from s.
// Only the rightmost annotation is stripped (one pass).
func stripTrailingAnnotation(s string) string {
	if !strings.HasSuffix(s, "}") {
		return s
	}
	idx := strings.LastIndex(s, "{")
	if idx < 0 {
		return s
	}
	return s[:idx]
}

// stripTrailingExponent removes a trailing signed or unsigned integer from s.
// Matches the regex /[+-]?[0-9]+$/ from the right.
func stripTrailingExponent(s string) string {
	n := len(s)
	if n == 0 {
		return s
	}
	i := n
	for i > 0 && s[i-1] >= '0' && s[i-1] <= '9' {
		i--
	}
	if i == n {
		// No trailing digits.
		return s
	}
	// Optional sign before digits.
	if i > 0 && (s[i-1] == '+' || s[i-1] == '-') {
		i--
	}
	return s[:i]
}

// sortedPrefixesLongestFirst returns the UCUM prefix keys sorted longest first
// so that binary prefixes like "Ki" and "Mi" win over single-char ones.
var cachedPrefixes []string

func sortedPrefixesLongestFirst() []string {
	if cachedPrefixes != nil {
		return cachedPrefixes
	}
	pfxs := make([]string, 0, len(ucumPrefixes))
	for p := range ucumPrefixes {
		pfxs = append(pfxs, p)
	}
	sort.Slice(pfxs, func(i, j int) bool {
		if len(pfxs[i]) != len(pfxs[j]) {
			return len(pfxs[i]) > len(pfxs[j])
		}
		return pfxs[i] < pfxs[j]
	})
	cachedPrefixes = pfxs
	return cachedPrefixes
}

// suggestUCUM returns a UCUM code suggestion for s if s (case-insensitively)
// matches a key in verboseToUCUM. Returns empty string if no suggestion found.
func suggestUCUM(s string) string {
	lower := strings.ToLower(s)
	return verboseToUCUM[lower]
}
