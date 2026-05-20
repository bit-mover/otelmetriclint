package rules

import (
	"fmt"
	"strings"
)

// UnitSuffix flags metric names whose final dot-segment is (or ends with
// `_<unit>`) a unit word like `duration`, `seconds`, `bytes`. Units belong
// in metric.WithUnit(...), not in the name — the Prometheus exporter
// appends the unit on translation, producing double-encoded names like
// `foo_duration_seconds`.
type UnitSuffix struct {
	// Forbidden is the list of disallowed final-segment tokens.
	Forbidden []string
}

// ID returns the rule's stable identifier.
func (UnitSuffix) ID() string { return "unit_suffix" }

// Check reports a diagnostic if call.Name's final dot-segment is or ends
// in `_<unit>` for any unit in Forbidden.
func (u UnitSuffix) Check(call MetricCall) []Diagnostic {
	if call.Name == "" || len(u.Forbidden) == 0 {
		return nil
	}
	finalSeg := call.Name
	if i := strings.LastIndex(call.Name, "."); i >= 0 {
		finalSeg = call.Name[i+1:]
	}
	for _, suffix := range u.Forbidden {
		if finalSeg == suffix {
			return []Diagnostic{u.diag(call, suffix)}
		}
		if strings.HasSuffix(finalSeg, "_"+suffix) {
			return []Diagnostic{u.diag(call, suffix)}
		}
	}
	return nil
}

func (u UnitSuffix) diag(call MetricCall, suffix string) Diagnostic {
	return Diagnostic{
		Pos: call.Pos,
		Message: fmt.Sprintf(
			"metric name %q has unit-like suffix %q — units belong in metric.WithUnit(...); the Prometheus exporter appends the unit on translation",
			call.Name, suffix,
		),
		RuleID: u.ID(),
	}
}
