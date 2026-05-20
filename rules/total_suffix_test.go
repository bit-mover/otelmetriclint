package rules

import "testing"

func TestTotalSuffixRule(t *testing.T) {
	r := TotalSuffix{}
	if r.ID() != "total_suffix" {
		t.Fatalf("ID = %q", r.ID())
	}
	// Counter ending in _total → diagnosed
	if got := r.Check(MetricCall{Name: "eventhub.drops_total", Kind: KindInt64Counter}); len(got) != 1 {
		t.Errorf("expected diagnostic, got %d", len(got))
	}
	// Float64Counter ending in _total → also diagnosed
	if got := r.Check(MetricCall{Name: "foo.bar_total", Kind: KindFloat64Counter}); len(got) != 1 {
		t.Errorf("expected diagnostic, got %d", len(got))
	}
	// Histogram ending in _total → NOT diagnosed (rule is counter-scoped)
	if got := r.Check(MetricCall{Name: "foo.bar_total", Kind: KindInt64Histogram}); len(got) != 0 {
		t.Errorf("rule should be counter-scoped, got %d for histogram", len(got))
	}
	// Counter not ending _total → clean
	if got := r.Check(MetricCall{Name: "foo.bar", Kind: KindInt64Counter}); len(got) != 0 {
		t.Errorf("clean counter flagged: %v", got)
	}
	// Empty name → noop
	if got := r.Check(MetricCall{Name: "", Kind: KindInt64Counter}); len(got) != 0 {
		t.Errorf("rule should skip empty name, got %v", got)
	}
}
