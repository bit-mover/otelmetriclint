package rules

import (
	"fmt"
	"strings"
)

// TotalSuffix flags OTel counters whose name ends in `_total`. The
// Prometheus exporter appends `_total` on its own, so a name ending in
// `_total` produces a double-suffixed Prom name like `foo_total_total`.
type TotalSuffix struct{}

// ID returns the rule's stable identifier.
func (TotalSuffix) ID() string { return "total_suffix" }

// Check reports a diagnostic when a counter name ends in `_total`.
func (t TotalSuffix) Check(call MetricCall) []Diagnostic {
	if call.Name == "" || !call.Kind.IsCounter() {
		return nil
	}
	if !strings.HasSuffix(call.Name, "_total") {
		return nil
	}
	return []Diagnostic{{
		Pos:     call.Pos,
		Message: fmt.Sprintf("counter name %q must not end in _total — the Prometheus exporter appends it on translation, producing a double-encoded name", call.Name),
		RuleID:  t.ID(),
	}}
}
