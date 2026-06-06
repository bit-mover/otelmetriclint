package rules

import (
	"strings"
	"testing"
)

func TestSemconvID(t *testing.T) {
	r := Semconv{}
	if r.ID() != "semconv" {
		t.Fatalf("ID = %q, want %q", r.ID(), "semconv")
	}
}

func TestSemconvCheck(t *testing.T) {
	cases := []struct {
		name        string
		metricName  string
		allow       []string
		wantCount   int
		wantSubstr  string   // non-empty: must appear in message
		wantAbsent  string   // non-empty: must NOT appear in message
	}{
		{
			name:       "empty name returns nil",
			metricName: "",
			wantCount:  0,
		},
		{
			name:       "unclaimed project root passes silently",
			metricName: "cedar.foo.bar",
			wantCount:  0,
		},
		{
			name:       "registered semconv name passes",
			metricName: "http.server.request.duration",
			wantCount:  0,
		},
		{
			name:       "another registered semconv name passes",
			metricName: "http.client.request.duration",
			wantCount:  0,
		},
		{
			name:       "claimed root with unregistered name emits diagnostic",
			metricName: "http.server.totally.unknown",
			wantCount:  1,
			wantSubstr: `"http"`,
		},
		{
			name:       "near-miss emits suggestion",
			metricName: "http.server.request.durations", // trailing s
			wantCount:  1,
			wantSubstr: `did you mean "http.server.request.duration"`,
		},
		{
			name:       "far-miss within claimed namespace has no suggestion",
			metricName: "http.completely.made.up.thing.really.long.name",
			wantCount:  1,
			wantAbsent: "did you mean",
		},
		{
			name:       "allow entry under claimed namespace suppresses diagnostic",
			metricName: "http.server.custom.metric",
			allow:      []string{"http.server.custom.metric"},
			wantCount:  0,
		},
		{
			name:       "diagnostic message contains namespace",
			metricName: "k8s.pod.totally.fake",
			wantCount:  1,
			wantSubstr: `"k8s"`,
		},
		{
			name:       "diagnostic message contains metric name",
			metricName: "http.server.bogus",
			wantCount:  1,
			wantSubstr: `"http.server.bogus"`,
		},
		{
			name:       "diagnostic ruleID is semconv",
			metricName: "http.server.bogus2",
			wantCount:  1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := Semconv{Allow: tc.allow}
			got := r.Check(MetricCall{Name: tc.metricName, Kind: KindInt64Counter})

			if len(got) != tc.wantCount {
				t.Fatalf("expected %d diagnostic(s), got %d: %v", tc.wantCount, len(got), got)
			}
			if tc.wantCount == 0 {
				return
			}
			diag := got[0]

			if diag.RuleID != "semconv" {
				t.Errorf("RuleID = %q, want %q", diag.RuleID, "semconv")
			}
			if tc.wantSubstr != "" && !strings.Contains(diag.Message, tc.wantSubstr) {
				t.Errorf("message should contain %q, got: %q", tc.wantSubstr, diag.Message)
			}
			if tc.wantAbsent != "" && strings.Contains(diag.Message, tc.wantAbsent) {
				t.Errorf("message should NOT contain %q, got: %q", tc.wantAbsent, diag.Message)
			}
		})
	}
}

func TestSemconvNearMissSuggestionBoundary(t *testing.T) {
	// Verify that near-miss (distance ≤ 3 AND < half name length) produces suggestion.
	r := Semconv{}

	// "http.server.request.durations" vs "http.server.request.duration" => distance 1
	// name length = 30, half = 15; 1 < 15, so suggestion expected.
	diags := r.Check(MetricCall{Name: "http.server.request.durations"})
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if !strings.Contains(diags[0].Message, "did you mean") {
		t.Errorf("expected 'did you mean' in message, got: %q", diags[0].Message)
	}
	if !strings.Contains(diags[0].Message, `"http.server.request.duration"`) {
		t.Errorf("expected suggestion to be http.server.request.duration, got: %q", diags[0].Message)
	}
}

func TestSemconvLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
	}
	for _, tc := range cases {
		got := levenshtein(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
