package rules

import "testing"

func TestPluralizationRule(t *testing.T) {
	t.Run("ID", func(t *testing.T) {
		r := Pluralization{}
		if r.ID() != "pluralization" {
			t.Fatalf("ID = %q", r.ID())
		}
	})

	// flagged: plural leaf on UpDownCounter → diagnostic expected
	flagged := []string{
		"db.client.connections",
		"http.server.requests",
		"system.users",
		"queue.messages",
		"db.queries",
		"log.entries",
		"background.jobs",
		"db.rows",
	}
	for _, name := range flagged {
		t.Run("flagged/"+name, func(t *testing.T) {
			r := Pluralization{}
			got := r.Check(MetricCall{Name: name, Kind: KindInt64UpDownCounter})
			if len(got) != 1 {
				t.Errorf("expected 1 diagnostic for %q, got %d", name, len(got))
			}
		})
	}

	// clean via suffix exception: ends in -ss, -us, or -is → not plural
	cleanSuffix := []string{
		"system.status",   // ends in "us"
		"worker.process",  // ends in "ss"
		"db.access",       // ends in "ss"
		"data.analysis",   // ends in "is"
		"pop.census",      // ends in "us"
	}
	for _, name := range cleanSuffix {
		t.Run("clean_suffix/"+name, func(t *testing.T) {
			r := Pluralization{}
			got := r.Check(MetricCall{Name: name, Kind: KindInt64UpDownCounter})
			if len(got) != 0 {
				t.Errorf("unexpected diagnostic for suffix-exempt %q: %v", name, got)
			}
		})
	}

	// clean via no trailing-s: these don't end in "s" at all
	cleanNoS := []string{
		"db.connection.count",
		"payload.size",
		"request.duration",
	}
	for _, name := range cleanNoS {
		t.Run("clean_no_s/"+name, func(t *testing.T) {
			r := Pluralization{}
			got := r.Check(MetricCall{Name: name, Kind: KindInt64UpDownCounter})
			if len(got) != 0 {
				t.Errorf("unexpected diagnostic for no-s leaf %q: %v", name, got)
			}
		})
	}

	// clean via length: leaf token is fewer than 4 chars → not flagged
	cleanLength := []string{
		"host.os",  // leaf "os" len=2
		"fd.fds",   // leaf "fds" len=3
	}
	for _, name := range cleanLength {
		t.Run("clean_length/"+name, func(t *testing.T) {
			r := Pluralization{}
			got := r.Check(MetricCall{Name: name, Kind: KindInt64UpDownCounter})
			if len(got) != 0 {
				t.Errorf("unexpected diagnostic for short leaf %q: %v", name, got)
			}
		})
	}

	// clean via allowlist: series/kubernetes pass only when explicitly in Allow
	allowlisted := []struct {
		name  string
		allow []string
	}{
		{"time.series", []string{"series"}},
		{"cluster.kubernetes", []string{"kubernetes"}},
	}
	for _, tc := range allowlisted {
		t.Run("clean_allowlist/"+tc.name, func(t *testing.T) {
			r := Pluralization{Allow: tc.allow}
			got := r.Check(MetricCall{Name: tc.name, Kind: KindInt64UpDownCounter})
			if len(got) != 0 {
				t.Errorf("unexpected diagnostic for allowlisted %q: %v", tc.name, got)
			}
		})
	}

	// without the allowlist entry the same names should flag
	t.Run("flagged_without_allowlist/series", func(t *testing.T) {
		r := Pluralization{}
		got := r.Check(MetricCall{Name: "time.series", Kind: KindInt64UpDownCounter})
		if len(got) != 1 {
			t.Errorf("expected diagnostic for series without allowlist, got %d", len(got))
		}
	})
	t.Run("flagged_without_allowlist/kubernetes", func(t *testing.T) {
		r := Pluralization{}
		got := r.Check(MetricCall{Name: "cluster.kubernetes", Kind: KindInt64UpDownCounter})
		if len(got) != 1 {
			t.Errorf("expected diagnostic for kubernetes without allowlist, got %d", len(got))
		}
	})

	// kind gate: plural leaf on non-UpDownCounter kinds → no diagnostic
	t.Run("kind_gate/Int64Counter", func(t *testing.T) {
		r := Pluralization{}
		got := r.Check(MetricCall{Name: "db.client.connections", Kind: KindInt64Counter})
		if len(got) != 0 {
			t.Errorf("expected no diagnostic for Int64Counter, got %d", len(got))
		}
	})
	t.Run("kind_gate/Int64Histogram", func(t *testing.T) {
		r := Pluralization{}
		got := r.Check(MetricCall{Name: "db.client.connections", Kind: KindInt64Histogram})
		if len(got) != 0 {
			t.Errorf("expected no diagnostic for Int64Histogram, got %d", len(got))
		}
	})

	// observable UpDownCounter variants also flag
	t.Run("observable_up_down/Int64", func(t *testing.T) {
		r := Pluralization{}
		got := r.Check(MetricCall{Name: "db.client.connections", Kind: KindInt64ObservableUpDownCounter})
		if len(got) != 1 {
			t.Errorf("expected 1 diagnostic for Int64ObservableUpDownCounter, got %d", len(got))
		}
	})
	t.Run("observable_up_down/Float64", func(t *testing.T) {
		r := Pluralization{}
		got := r.Check(MetricCall{Name: "db.client.connections", Kind: KindFloat64ObservableUpDownCounter})
		if len(got) != 1 {
			t.Errorf("expected 1 diagnostic for Float64ObservableUpDownCounter, got %d", len(got))
		}
	})

	// empty name → no diagnostic
	t.Run("empty_name", func(t *testing.T) {
		r := Pluralization{}
		got := r.Check(MetricCall{Name: "", Kind: KindInt64UpDownCounter})
		if len(got) != 0 {
			t.Errorf("expected no diagnostic for empty name, got %v", got)
		}
	})

	// user-supplied Allow entry suppresses an otherwise-flagged leaf
	t.Run("user_allow_suppresses", func(t *testing.T) {
		r := Pluralization{Allow: []string{"connections"}}
		got := r.Check(MetricCall{Name: "db.client.connections", Kind: KindInt64UpDownCounter})
		if len(got) != 0 {
			t.Errorf("expected no diagnostic when leaf is in Allow, got %v", got)
		}
	})
}
