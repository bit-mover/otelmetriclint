package rules

import "testing"

func TestStructuralRule(t *testing.T) {
	r := Structural{}
	if r.ID() != "structural" {
		t.Fatalf("ID = %q", r.ID())
	}
	cases := []struct {
		name string
		want bool // true if it should be diagnosed
	}{
		{"iam.users.registrations", false},
		{"cedar.authorize", false},
		{"a", false},
		{"Foo.bar", true},        // uppercase
		{"9foo", true},           // leading digit
		{"foo..bar", true},       // double dot delimiter
		{"foo__bar", true},       // double underscore inside a segment
		{"foo.9bar", true},       // segment starts with digit
		{"foo.bar_", false},      // trailing underscore is allowed by `[a-z][a-z0-9_]*`
		{"-foo", true},           // non-allowed char
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

func TestStructuralRuleSkipsEmptyName(t *testing.T) {
	// Empty Name means recognizer failed to resolve — string_literal rule
	// handles that, structural stays quiet to avoid duplicate noise.
	if got := (Structural{}).Check(MetricCall{Name: ""}); len(got) != 0 {
		t.Errorf("structural should not fire on empty name, got %v", got)
	}
}
