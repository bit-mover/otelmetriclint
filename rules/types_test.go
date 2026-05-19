package rules

import "testing"

func TestInstrumentKindString(t *testing.T) {
	cases := []struct {
		kind InstrumentKind
		want string
	}{
		{KindInt64Counter, "Int64Counter"},
		{KindFloat64Counter, "Float64Counter"},
		{KindInt64UpDownCounter, "Int64UpDownCounter"},
		{KindFloat64UpDownCounter, "Float64UpDownCounter"},
		{KindInt64Histogram, "Int64Histogram"},
		{KindFloat64Histogram, "Float64Histogram"},
		{KindInt64Gauge, "Int64Gauge"},
		{KindFloat64Gauge, "Float64Gauge"},
	}
	for _, tc := range cases {
		if got := tc.kind.String(); got != tc.want {
			t.Errorf("InstrumentKind(%d).String() = %q, want %q", tc.kind, got, tc.want)
		}
	}
}

func TestRegistryEnabledRulesRespectsConfig(t *testing.T) {
	reg := Registry{Rules: []Rule{
		stubRule{id: "alpha"},
		stubRule{id: "beta"},
		stubRule{id: "gamma"},
	}}
	enabled := reg.Enabled(map[string]bool{"alpha": true, "beta": false})
	if len(enabled) != 2 || enabled[0].ID() != "alpha" || enabled[1].ID() != "gamma" {
		t.Errorf("unexpected enabled set: %v", ruleIDs(enabled))
	}
}

func ruleIDs(rs []Rule) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.ID()
	}
	return out
}

type stubRule struct{ id string }

func (s stubRule) ID() string                    { return s.id }
func (s stubRule) Check(MetricCall) []Diagnostic { return nil }
