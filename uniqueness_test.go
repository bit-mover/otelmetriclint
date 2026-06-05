package otelmetriclint

import (
	"testing"
)

// TestRegisteredNamesString covers the String() method of the registeredNames fact.
func TestRegisteredNamesString(t *testing.T) {
	tests := []struct {
		name string
		fact *registeredNames
		want string
	}{
		{
			name: "empty",
			fact: &registeredNames{},
			want: "registeredNames{}",
		},
		{
			name: "single entry",
			fact: &registeredNames{
				Names: []registeredName{
					{Name: "foo.bar", Pos: "a.go:10:14"},
				},
			},
			want: "registeredNames{foo.bar@a.go:10:14}",
		},
		{
			name: "multiple entries",
			fact: &registeredNames{
				Names: []registeredName{
					{Name: "foo.bar", Pos: "a.go:10:14"},
					{Name: "baz.qux", Pos: "a.go:11:14"},
				},
			},
			want: "registeredNames{foo.bar@a.go:10:14, baz.qux@a.go:11:14}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fact.String()
			if got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestRegisteredNamesAFact ensures registeredNames satisfies the analysis.Fact
// interface (AFact method exists and does not panic).
func TestRegisteredNamesAFact(_ *testing.T) {
	rn := &registeredNames{}
	// AFact is a marker; calling it should not panic.
	rn.AFact()
}
