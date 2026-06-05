package rules

import (
	"fmt"
	"strings"
)

// Pluralization flags UpDownCounter (and ObservableUpDownCounter) names whose
// leaf token appears pluralized. UpDownCounters represent a current value, not
// a count of discrete events, and should use a singular noun.
type Pluralization struct {
	// Allow is the set of lowercase leaf tokens that are exempt from the
	// pluralization heuristic. The built-in list is supplied by analyzer.go
	// via this field; users may append additional entries through config.
	Allow []string
}

// ID returns the rule's stable identifier.
func (Pluralization) ID() string { return "pluralization" }

// Check reports a diagnostic when an UpDownCounter name's leaf token looks
// pluralized and is not in the allowlist.
func (p Pluralization) Check(call MetricCall) []Diagnostic {
	if call.Name == "" || !call.Kind.IsUpDownCounter() {
		return nil
	}

	// Build the allowlist map once per Check call. The Allow slice is small
	// and this approach keeps the rule stateless (no Init method needed).
	allowed := make(map[string]struct{}, len(p.Allow))
	for _, s := range p.Allow {
		allowed[strings.ToLower(s)] = struct{}{}
	}

	leaf := leafToken(call.Name)
	if leaf == "" {
		return nil
	}

	if !looksPlural(leaf, allowed) {
		return nil
	}

	return []Diagnostic{{
		Pos: call.Pos,
		Message: fmt.Sprintf(
			"UpDownCounter name %q appears pluralized — UpDownCounters represent a current value, not a count of discrete events; prefer a singular form (e.g. %q)",
			call.Name,
			singularHint(call.Name, leaf),
		),
		RuleID: p.ID(),
	}}
}

// leafToken extracts the last `_`-separated token of the last `.`-segment of
// name, lowercased. For example "db.client.connections" → "connections",
// "http.server.request_count" → "count".
func leafToken(name string) string {
	// Take the last dot-segment.
	seg := name
	if i := strings.LastIndex(name, "."); i >= 0 {
		seg = name[i+1:]
	}
	// Take the last underscore-token within that segment.
	token := seg
	if i := strings.LastIndex(seg, "_"); i >= 0 {
		token = seg[i+1:]
	}
	return strings.ToLower(token)
}

// looksPlural returns true when leaf appears to be an English plural that the
// heuristic would flag. allowed is the combined built-in + user allowlist.
func looksPlural(leaf string, allowed map[string]struct{}) bool {
	if _, ok := allowed[leaf]; ok {
		return false
	}
	if !strings.HasSuffix(leaf, "s") {
		return false
	}
	if strings.HasSuffix(leaf, "ss") || strings.HasSuffix(leaf, "us") || strings.HasSuffix(leaf, "is") {
		return false
	}
	return len(leaf) >= 4
}

// singularHint builds a human-readable example of a singular form for the
// diagnostic message. It replaces the leaf (last dot-segment) with
// "<singular>.count" where <singular> strips the trailing `s` from the leaf
// segment — intentionally rough, just for illustration.
func singularHint(name, leaf string) string {
	// Find the last dot-segment boundary.
	prefix := ""
	seg := name
	if i := strings.LastIndex(name, "."); i >= 0 {
		prefix = name[:i+1]
		seg = name[i+1:]
	}

	// Strip the trailing `s` from the segment (not just the leaf token, so
	// "connections" → "connection") for the suggestion.
	singular := seg
	if strings.HasSuffix(strings.ToLower(seg), leaf) && strings.HasSuffix(leaf, "s") {
		singular = seg[:len(seg)-1]
	}

	return prefix + singular + ".count"
}
