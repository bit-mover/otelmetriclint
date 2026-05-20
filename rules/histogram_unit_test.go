package rules

import (
	"go/ast"
	"go/parser"
	"testing"
)

func TestHistogramUnitRule(t *testing.T) {
	r := HistogramUnit{}
	if r.ID() != "histogram_unit" {
		t.Fatalf("ID = %q", r.ID())
	}
	// Helper to build a fake option call expression
	mkOpt := func(src string) ast.Expr {
		expr, err := parser.ParseExpr(src)
		if err != nil {
			t.Fatalf("parse %q: %v", src, err)
		}
		return expr
	}
	// Histogram with no options → diagnosed
	if got := r.Check(MetricCall{Name: "foo.bar", Kind: KindInt64Histogram, Options: nil}); len(got) != 1 {
		t.Errorf("expected diagnostic, got %d", len(got))
	}
	// Histogram with metric.WithUnit("s") → clean
	mc := MetricCall{
		Name:    "foo.bar",
		Kind:    KindFloat64Histogram,
		Options: []ast.Expr{mkOpt(`metric.WithUnit("s")`)},
	}
	if got := r.Check(mc); len(got) != 0 {
		t.Errorf("histogram with WithUnit flagged: %v", got)
	}
	// Histogram with WithDescription but no WithUnit → diagnosed
	mc = MetricCall{
		Name:    "foo.bar",
		Kind:    KindInt64Histogram,
		Options: []ast.Expr{mkOpt(`metric.WithDescription("hi")`)},
	}
	if got := r.Check(mc); len(got) != 1 {
		t.Errorf("histogram with only WithDescription should be diagnosed, got %d", len(got))
	}
	// Non-histogram with no WithUnit → clean (rule is histogram-scoped)
	if got := r.Check(MetricCall{Name: "foo.bar", Kind: KindInt64Counter}); len(got) != 0 {
		t.Errorf("non-histogram flagged: %v", got)
	}
	// Empty name → noop
	if got := r.Check(MetricCall{Name: "", Kind: KindInt64Histogram}); len(got) != 0 {
		t.Errorf("empty name flagged: %v", got)
	}
}
