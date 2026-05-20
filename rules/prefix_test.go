package rules

import "testing"

func TestPrefixRule(t *testing.T) {
	r := Prefix{Allowed: []string{"iam", "customers"}}
	if r.ID() != "prefix" {
		t.Fatalf("ID = %q", r.ID())
	}
	cases := []struct {
		name string
		want bool // true if diagnosed
	}{
		{"iam.users.registrations", false},
		{"customers.groups.create", false},
		{"rogue.thing", true},
		{"ops.things", true}, // not in allowlist
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

func TestPrefixRuleEmptyAllowlistIsNoop(t *testing.T) {
	r := Prefix{}
	if got := r.Check(MetricCall{Name: "rogue.thing"}); len(got) != 0 {
		t.Errorf("empty allowlist should pass everything, got %v", got)
	}
}

func TestPrefixRuleSkipsEmptyName(t *testing.T) {
	r := Prefix{Allowed: []string{"iam"}}
	if got := r.Check(MetricCall{Name: ""}); len(got) != 0 {
		t.Errorf("prefix should not fire on empty name, got %v", got)
	}
}
