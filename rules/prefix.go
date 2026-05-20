package rules

import (
	"fmt"
	"strings"
)

// Prefix enforces that the first dot-segment of a metric name is in an
// allowlist. An empty allowlist disables the rule.
type Prefix struct {
	Allowed []string
}

// ID returns the rule's stable identifier.
func (Prefix) ID() string { return "prefix" }

// Check reports a diagnostic when the first dot-segment isn't in the allowlist.
func (p Prefix) Check(call MetricCall) []Diagnostic {
	if call.Name == "" || len(p.Allowed) == 0 {
		return nil
	}
	first, _, _ := strings.Cut(call.Name, ".")
	if first == "" {
		first = call.Name
	}
	for _, allowed := range p.Allowed {
		if first == allowed {
			return nil
		}
	}
	return []Diagnostic{{
		Pos:     call.Pos,
		Message: fmt.Sprintf("metric name %q has prefix %q not in the allowlist %v", call.Name, first, p.Allowed),
		RuleID:  "prefix",
	}}
}
