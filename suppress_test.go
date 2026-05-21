package otelmetriclint

import "testing"

func TestMatchesNoLint(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		// Positive — bare nolint suppresses everything.
		{"//nolint", true},
		{"//nolint ", true},
		{"//nolint\t", true},

		// Positive — linter list including otelmetriclint.
		{"//nolint:otelmetriclint", true},
		{"//nolint:otelmetriclint,gocritic", true},
		{"//nolint:gocritic,otelmetriclint", true},
		{"//nolint:gocritic, otelmetriclint", true},
		{"//nolint:otelmetriclint ", true},
		{"//nolint:otelmetriclint // legacy migration", true},
		{"/*nolint:otelmetriclint*/", true},
		{"/* nolint:otelmetriclint */", true},

		// Negative — wrong shape.
		{"// nolint:otelmetriclint", false}, // golangci's space-after-// quirk
		{"//NOLINT:otelmetriclint", false},  // wrong case
		{"//nolint:gocritic", false},
		{"//nolint:otelmetriclintfoo", false}, // not exact-name
		{"//notanolint:otelmetriclint", false},
		{"//nolintfoo", false},
		{"// just a comment", false},
		{"", false},
		// "//nolint:" with an empty list is NOT treated as bare //nolint;
		// it matches nothing. Mirrors golangci-lint.
		{"//nolint:", false},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			if got := matchesNoLint(c.in); got != c.want {
				t.Errorf("matchesNoLint(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}
