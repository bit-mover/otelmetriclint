package rules

import "testing"

func TestUnitSuffixRule(t *testing.T) {
	r := UnitSuffix{Forbidden: defaultUnitSuffixList()}
	if r.ID() != "unit_suffix" {
		t.Fatalf("ID = %q", r.ID())
	}
	cases := []struct {
		name string
		want bool
	}{
		// final segment is a unit word
		{"cedar.authorize.duration", true},
		{"foo.seconds", true},
		{"foo.bytes", true},
		{"foo.ms", true},
		// final segment ends in _unit (load_duration, batch_duration)
		{"es.aggregate_store.load_duration", true},
		{"outbox.batch_duration", true},
		// not a unit
		{"iam.users.registrations", false},
		{"cedar.authorize", false},
		// empty name handled elsewhere
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := r.Check(MetricCall{Name: tc.name, Kind: KindInt64Counter})
			if tc.want && len(got) == 0 {
				t.Errorf("expected diagnostic for %q", tc.name)
			}
			if !tc.want && len(got) != 0 {
				t.Errorf("unexpected diagnostic for %q: %v", tc.name, got)
			}
		})
	}
}

// defaultUnitSuffixList is exported to tests for reuse; same list lives
// in the analyzer's config Default() in the root package.
func defaultUnitSuffixList() []string {
	return []string{"duration", "seconds", "bytes", "ms", "us", "ns", "s", "kb", "mb", "gb", "b", "count", "total"}
}
