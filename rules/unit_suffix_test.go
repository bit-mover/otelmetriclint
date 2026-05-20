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
		// final segment is a UCUM unit code or its expansion
		{"foo.seconds", true},
		{"foo.bytes", true},
		{"foo.ms", true},
		// `duration` and `count` are quantity descriptors, not units —
		// OTel semconv uses them canonically (http.server.request.duration,
		// db.client.connection.count) — the rule must NOT flag them.
		{"http.server.request.duration", false},
		{"db.client.connection.count", false},
		{"cedar.authorize.duration", false},
		// `_<unit>` as a final-segment suffix still fires for true units
		{"foo.bar_seconds", true},
		{"foo.bar_bytes", true},
		// but not for quantity descriptors
		{"es.aggregate_store.load_duration", false},
		{"outbox.batch_duration", false},
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

// defaultUnitSuffixList mirrors the list in the analyzer's config
// Default() in the root package. Contains UCUM unit codes and their
// expansions only — quantity descriptors like `duration` and `count` are
// excluded because OTel semconv uses them canonically.
func defaultUnitSuffixList() []string {
	return []string{"seconds", "bytes", "ms", "us", "ns", "s", "kb", "mb", "gb", "b", "total"}
}
