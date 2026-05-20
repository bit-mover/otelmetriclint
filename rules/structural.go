package rules

import (
	"fmt"
	"regexp"
	"strings"
)

// Structural enforces the syntactic shape of an OTel metric name:
// - ASCII lowercase letters, digits, underscores, dots only
// - Must start with a letter
// - No `..` or `__` (consecutive delimiters)
// - Each dot-segment matches ^[a-z][a-z0-9_]*$
//
// It does NOT enforce English-word boundary snake_case (no dictionary).
// "users.getuser" passes structural — it's still snake_case-shaped, just
// missing an internal underscore the author meant to include.
type Structural struct{}

// ID returns the rule's stable identifier.
func (Structural) ID() string { return "structural" }

var (
	overallShape = regexp.MustCompile(`^[a-z][a-z0-9._]*$`)
	segmentShape = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

// Check enforces the syntactic shape of a metric name.
func (s Structural) Check(call MetricCall) []Diagnostic {
	if call.Name == "" {
		return nil // handled by string_literal
	}
	name := call.Name
	if !overallShape.MatchString(name) {
		return []Diagnostic{{
			Pos:     call.Pos,
			Message: fmt.Sprintf("metric name %q must be ASCII-lowercase letters, digits, dots, or underscores, starting with a letter", name),
			RuleID:  s.ID(),
		}}
	}
	if strings.Contains(name, "..") {
		return []Diagnostic{{
			Pos:     call.Pos,
			Message: fmt.Sprintf("metric name %q contains consecutive dot delimiters", name),
			RuleID:  s.ID(),
		}}
	}
	if strings.Contains(name, "__") {
		return []Diagnostic{{
			Pos:     call.Pos,
			Message: fmt.Sprintf("metric name %q contains consecutive underscore delimiters", name),
			RuleID:  s.ID(),
		}}
	}
	for _, seg := range strings.Split(name, ".") {
		if !segmentShape.MatchString(seg) {
			return []Diagnostic{{
				Pos:     call.Pos,
				Message: fmt.Sprintf("metric name %q has invalid segment %q — must start with a letter and contain only lowercase letters, digits, and underscores", name, seg),
				RuleID:  s.ID(),
			}}
		}
	}
	return nil
}
