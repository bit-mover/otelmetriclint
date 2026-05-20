package rules

import (
	"fmt"
	"go/ast"
)

// HistogramUnit fires when a histogram is created without a
// metric.WithUnit(...) option. Once unit_suffix forbids units in the
// name, the unit must live somewhere — this rule enforces that.
//
// Matching is selector-name based ("WithUnit") rather than fully-typed
// because Rule.Check receives only the MetricCall value. The recognizer
// already established this call site creates an OTel histogram, so a
// name match on the option is sufficient — the worst case is a false
// negative when a third-party WithUnit happens to share the name (rare).
type HistogramUnit struct{}

// ID returns the rule's stable identifier.
func (HistogramUnit) ID() string { return "histogram_unit" }

// Check reports a diagnostic when a histogram has no WithUnit option.
func (h HistogramUnit) Check(call MetricCall) []Diagnostic {
	if call.Name == "" || !call.Kind.IsHistogram() {
		return nil
	}
	for _, opt := range call.Options {
		if isWithUnitCall(opt) {
			return nil
		}
	}
	return []Diagnostic{{
		Pos: call.Pos,
		Message: fmt.Sprintf(
			"histogram %q must set metric.WithUnit(...) — units belong in WithUnit, not in the metric name",
			call.Name,
		),
		RuleID: h.ID(),
	}}
}

// isWithUnitCall reports whether opt is a call to a function named
// WithUnit (typically metric.WithUnit). Conservatively name-based — the
// recognizer narrows further by establishing this is an OTel histogram.
func isWithUnitCall(opt ast.Expr) bool {
	call, ok := opt.(*ast.CallExpr)
	if !ok {
		return false
	}
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		return fun.Sel.Name == "WithUnit"
	case *ast.Ident:
		return fun.Name == "WithUnit"
	}
	return false
}
