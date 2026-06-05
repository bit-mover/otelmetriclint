package rules

import (
	"go/ast"
	"go/parser"
	"testing"
)

func TestUCUMUnitID(t *testing.T) {
	r := UCUMUnit{}
	if r.ID() != "ucum_unit" {
		t.Fatalf("ID = %q, want %q", r.ID(), "ucum_unit")
	}
}

func TestValidUCUM(t *testing.T) {
	valid := []string{
		"s",          // second — base atom
		"ms",         // milli + second
		"km",         // kilo + metre
		"By",         // byte — atom
		"KiBy",       // kibi + byte — binary prefix
		"MHz",        // mega + hertz
		"m2",         // metre squared — exponent
		"Cel",        // degree Celsius — atom
		"%",          // percent — atom
		"1",          // unity — atom
		"{event}",    // whole-string annotation
		"By{transmitted}", // atom + trailing annotation
		"By/s",       // byte per second — compound
		"kg.m/s2",    // compound: kilo+gram · metre / second²
		"mol/L",      // moles per litre
		"{packet}/s", // bare annotation term in a compound (annotated 1/s)
		"1{cell}",    // unity with a trailing annotation
	}
	for _, s := range valid {
		if !validUCUM(s) {
			t.Errorf("validUCUM(%q) = false, want true", s)
		}
	}

	invalid := []string{
		"seconds",  // verbose English — not a UCUM atom
		"bytes",    // verbose English — not a UCUM atom
		"foobar",   // made-up
		"garbage",  // made-up
	}
	for _, s := range invalid {
		if validUCUM(s) {
			t.Errorf("validUCUM(%q) = true, want false", s)
		}
	}
}

func TestSuggestUCUM(t *testing.T) {
	hits := map[string]string{
		"seconds":      "s",
		"Seconds":      "s",   // case-insensitive
		"SECONDS":      "s",
		"bytes":        "By",
		"milliseconds": "ms",
		"percent":      "%",
	}
	for input, want := range hits {
		if got := suggestUCUM(input); got != want {
			t.Errorf("suggestUCUM(%q) = %q, want %q", input, got, want)
		}
	}

	misses := []string{"foobar", "s", "By", ""}
	for _, input := range misses {
		if got := suggestUCUM(input); got != "" {
			t.Errorf("suggestUCUM(%q) = %q, want empty", input, got)
		}
	}
}

func TestUCUMUnitCheck(t *testing.T) {
	r := UCUMUnit{}

	mkOpt := func(src string) ast.Expr {
		expr, err := parser.ParseExpr(src)
		if err != nil {
			t.Fatalf("parse %q: %v", src, err)
		}
		return expr
	}

	t.Run("valid unit no diagnostic", func(t *testing.T) {
		mc := MetricCall{
			Name:    "foo.bar",
			Kind:    KindInt64Counter,
			Options: []ast.Expr{mkOpt(`metric.WithUnit("s")`)},
		}
		if got := r.Check(mc); len(got) != 0 {
			t.Errorf("expected no diagnostic, got %d: %v", len(got), got)
		}
	})

	t.Run("invalid unit with suggestion emits diagnostic", func(t *testing.T) {
		mc := MetricCall{
			Name:    "foo.bar",
			Kind:    KindInt64Counter,
			Options: []ast.Expr{mkOpt(`metric.WithUnit("seconds")`)},
		}
		got := r.Check(mc)
		if len(got) != 1 {
			t.Fatalf("expected 1 diagnostic, got %d", len(got))
		}
		if got[0].RuleID != "ucum_unit" {
			t.Errorf("RuleID = %q, want ucum_unit", got[0].RuleID)
		}
		// Should contain suggestion "s"
		if !containsSubstring(got[0].Message, `"s"`) {
			t.Errorf("message should mention suggestion: %q", got[0].Message)
		}
	})

	t.Run("invalid unit without suggestion emits generic diagnostic", func(t *testing.T) {
		mc := MetricCall{
			Name:    "foo.bar",
			Kind:    KindFloat64Histogram,
			Options: []ast.Expr{mkOpt(`metric.WithUnit("foobar")`)},
		}
		got := r.Check(mc)
		if len(got) != 1 {
			t.Fatalf("expected 1 diagnostic, got %d", len(got))
		}
		if containsSubstring(got[0].Message, "did you mean") {
			t.Errorf("generic message should not contain 'did you mean': %q", got[0].Message)
		}
		if !containsSubstring(got[0].Message, "ucum.org") {
			t.Errorf("generic message should link to ucum.org: %q", got[0].Message)
		}
	})

	t.Run("allow list short-circuits validation", func(t *testing.T) {
		r2 := UCUMUnit{Allow: []string{"custom_unit"}}
		mc := MetricCall{
			Name:    "foo.bar",
			Kind:    KindInt64Gauge,
			Options: []ast.Expr{mkOpt(`metric.WithUnit("custom_unit")`)},
		}
		if got := r2.Check(mc); len(got) != 0 {
			t.Errorf("allow-listed unit should not be flagged, got %d diagnostics", len(got))
		}
	})

	t.Run("non-literal argument skipped silently", func(t *testing.T) {
		mc := MetricCall{
			Name:    "foo.bar",
			Kind:    KindInt64Counter,
			Options: []ast.Expr{mkOpt(`metric.WithUnit(unitVar)`)},
		}
		if got := r.Check(mc); len(got) != 0 {
			t.Errorf("non-literal should be skipped, got %d diagnostics", len(got))
		}
	})

	t.Run("no WithUnit option no diagnostic", func(t *testing.T) {
		mc := MetricCall{
			Name:    "foo.bar",
			Kind:    KindInt64Counter,
			Options: []ast.Expr{mkOpt(`metric.WithDescription("hi")`)},
		}
		if got := r.Check(mc); len(got) != 0 {
			t.Errorf("no WithUnit should produce no diagnostic, got %d", len(got))
		}
	})

	t.Run("applies to all instrument kinds", func(t *testing.T) {
		kinds := []InstrumentKind{
			KindInt64Counter, KindFloat64Counter,
			KindInt64UpDownCounter, KindFloat64UpDownCounter,
			KindInt64Histogram, KindFloat64Histogram,
			KindInt64Gauge, KindFloat64Gauge,
		}
		for _, kind := range kinds {
			mc := MetricCall{
				Name:    "foo.bar",
				Kind:    kind,
				Options: []ast.Expr{mkOpt(`metric.WithUnit("garbage")`)},
			}
			got := r.Check(mc)
			if len(got) != 1 {
				t.Errorf("kind %v: expected 1 diagnostic, got %d", kind, len(got))
			}
		}
	})

	t.Run("compound unit is valid", func(t *testing.T) {
		mc := MetricCall{
			Name:    "foo.bar",
			Kind:    KindFloat64Histogram,
			Options: []ast.Expr{mkOpt(`metric.WithUnit("By/s")`)},
		}
		if got := r.Check(mc); len(got) != 0 {
			t.Errorf("By/s should be valid UCUM, got %d diagnostics", len(got))
		}
	})

	t.Run("annotation-only unit is valid", func(t *testing.T) {
		mc := MetricCall{
			Name:    "foo.bar",
			Kind:    KindInt64Counter,
			Options: []ast.Expr{mkOpt(`metric.WithUnit("{event}")`)},
		}
		if got := r.Check(mc); len(got) != 0 {
			t.Errorf("{event} should be valid UCUM, got %d diagnostics", len(got))
		}
	})
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
