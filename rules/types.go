// Package rules defines the rule interface, the call-site value type the
// recognizer produces (MetricCall), the diagnostic value type rules emit,
// and a Registry that orders and filters rules by enable state.
package rules

import (
	"go/ast"
	"go/token"
)

// InstrumentKind identifies which OTel metric instrument is being created.
type InstrumentKind int

// Instrument-kind constants identify the OTel metric instrument families
// the recognizer can match.
const (
	KindUnknown InstrumentKind = iota
	KindInt64Counter
	KindFloat64Counter
	KindInt64UpDownCounter
	KindFloat64UpDownCounter
	KindInt64Histogram
	KindFloat64Histogram
	KindInt64Gauge
	KindFloat64Gauge
)

func (k InstrumentKind) String() string {
	switch k {
	case KindInt64Counter:
		return "Int64Counter"
	case KindFloat64Counter:
		return "Float64Counter"
	case KindInt64UpDownCounter:
		return "Int64UpDownCounter"
	case KindFloat64UpDownCounter:
		return "Float64UpDownCounter"
	case KindInt64Histogram:
		return "Int64Histogram"
	case KindFloat64Histogram:
		return "Float64Histogram"
	case KindInt64Gauge:
		return "Int64Gauge"
	case KindFloat64Gauge:
		return "Float64Gauge"
	}
	return "Unknown"
}

// IsCounter reports whether the kind is one of the counter families.
func (k InstrumentKind) IsCounter() bool {
	return k == KindInt64Counter || k == KindFloat64Counter
}

// IsHistogram reports whether the kind is one of the histogram families.
func (k InstrumentKind) IsHistogram() bool {
	return k == KindInt64Histogram || k == KindFloat64Histogram
}

// MetricCall is a recognized metric-creation call site.
type MetricCall struct {
	Pos     token.Pos
	Kind    InstrumentKind
	Name    string
	NameArg ast.Expr
	Options []ast.Expr
	// EnclosingFuncs is the chain of enclosing *ast.FuncDecl or
	// *ast.FuncLit nodes, outermost first. Empty for top-level calls.
	// Populated by the recognizer; consumed by the suppression layer
	// to find directives placed above an enclosing func.
	EnclosingFuncs []ast.Node
}

// Diagnostic is one rule violation.
type Diagnostic struct {
	Pos     token.Pos
	Message string
	RuleID  string
}

// Rule checks one aspect of a MetricCall and returns zero or more diagnostics.
type Rule interface {
	ID() string
	Check(call MetricCall) []Diagnostic
}

// Registry holds the canonical, ordered set of rules.
type Registry struct {
	Rules []Rule
}

// Enabled returns the subset of registered rules whose ID maps to true in
// the enabled map. Missing keys default to true.
func (r Registry) Enabled(enabled map[string]bool) []Rule {
	out := make([]Rule, 0, len(r.Rules))
	for _, rule := range r.Rules {
		if v, ok := enabled[rule.ID()]; ok && !v {
			continue
		}
		out = append(out, rule)
	}
	return out
}
