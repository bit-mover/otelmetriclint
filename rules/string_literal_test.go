package rules

import (
	"go/token"
	"testing"
)

func TestStringLiteralRule(t *testing.T) {
	r := StringLiteral{}
	if r.ID() != "string_literal" {
		t.Fatalf("ID = %q", r.ID())
	}
	// Empty Name signals the recognizer couldn't resolve a literal.
	d := r.Check(MetricCall{Pos: token.Pos(42), Kind: KindInt64Counter, Name: ""})
	if len(d) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(d))
	}
	if d[0].RuleID != "string_literal" {
		t.Errorf("RuleID = %q", d[0].RuleID)
	}
	if d[0].Pos != token.Pos(42) {
		t.Errorf("Pos lost — got %v, want 42", d[0].Pos)
	}
	// Non-empty name = nothing to report.
	if got := r.Check(MetricCall{Kind: KindInt64Counter, Name: "foo.bar"}); len(got) != 0 {
		t.Errorf("expected no diagnostic for valid name, got %d", len(got))
	}
}
