package rules

import (
	"fmt"
	"strings"
)

// Semconv fires when a metric name belongs to a known OTel semantic-convention
// namespace (root segment before the first '.') but is not a registered semconv
// metric name.
//
// The Allow field is an optional per-team allowlist of non-registered names
// that should be accepted silently. Entries are matched against the full
// metric name.
type Semconv struct {
	Allow []string
}

// ID returns the rule's stable identifier.
func (Semconv) ID() string { return "semconv" }

// Check reports a diagnostic when the metric name's root segment is a known
// semconv namespace but the full name is not a registered semconv metric.
//
// Algorithm:
//  1. Empty name → nil.
//  2. root = segment before first '.'.
//  3. root not in semconvNamespaces → nil (false-positive firewall).
//  4. name in semconvMetricNames → nil.
//  5. name in effective Allow → nil.
//  6. Emit one diagnostic; append "did you mean?" when the closest registered
//     name within the same namespace is within Levenshtein distance ≤ 3 AND
//     strictly less than half the name's length.
func (s Semconv) Check(call MetricCall) []Diagnostic {
	name := call.Name
	// Step 1: empty name.
	if name == "" {
		return nil
	}

	// Step 2: extract root namespace.
	root, _, _ := strings.Cut(name, ".")

	// Step 3: false-positive firewall — only check known semconv namespaces.
	if !semconvNamespaces[root] {
		return nil
	}

	// Step 4: name is a registered semconv metric.
	if semconvMetricNames[name] {
		return nil
	}

	// Step 5: name is in the effective allow list.
	for _, a := range s.Allow {
		if name == a {
			return nil
		}
	}

	// Step 6: emit diagnostic, optionally with a "did you mean?" suggestion.
	msg := fmt.Sprintf(
		"metric name %q is not a registered OpenTelemetry semantic-convention metric in the %q namespace",
		name, root,
	)

	// Compute closest name in the same namespace.
	if suggestion, dist := closestSemconvName(name, root); suggestion != "" {
		threshold := len(name) / 2 // strictly less than half the name length
		if dist <= 3 && dist < threshold {
			msg += fmt.Sprintf(`; did you mean %q?`, suggestion)
		}
	}

	return []Diagnostic{{
		Pos:     call.Pos,
		Message: msg,
		RuleID:  s.ID(),
	}}
}

// closestSemconvName returns the registered semconv metric name in the given
// namespace that has the smallest Levenshtein distance to name, along with
// that distance. Returns ("", 0) if no candidates exist.
func closestSemconvName(name, namespace string) (string, int) {
	prefix := namespace + "."
	best := ""
	bestDist := -1

	for candidate := range semconvMetricNames {
		if !strings.HasPrefix(candidate, prefix) {
			continue
		}
		d := levenshtein(name, candidate)
		if bestDist < 0 || d < bestDist {
			bestDist = d
			best = candidate
		}
	}
	if bestDist < 0 {
		return "", 0
	}
	return best, bestDist
}

// levenshtein computes the standard Levenshtein edit distance between a and b.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Use two rows to save memory.
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			min := del
			if ins < min {
				min = ins
			}
			if sub < min {
				min = sub
			}
			curr[j] = min
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}
