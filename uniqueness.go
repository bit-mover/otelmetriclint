package otelmetriclint

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bit-mover/otelmetriclint/rules"
)

// registeredName is one metric-name registration from a single package,
// stored inside a registeredNames fact. Pos is the human-readable position
// string (pass.Fset.Position(call.Pos).String()), never a raw token.Pos.
type registeredName struct {
	Name string
	Pos  string
}

// registeredNames is a package fact that carries every metric name registered
// by a package. It is gob-encodable because its fields are all plain strings.
type registeredNames struct {
	Names []registeredName
}

// AFact marks registeredNames as a package fact for the analysis framework.
func (*registeredNames) AFact() {}

// String returns a human-readable representation of the fact.
func (r *registeredNames) String() string {
	parts := make([]string, len(r.Names))
	for i, n := range r.Names {
		parts[i] = fmt.Sprintf("%s@%s", n.Name, n.Pos)
	}
	return "registeredNames{" + strings.Join(parts, ", ") + "}"
}

// checkUniqueness detects metric names that are registered in more than one
// package. It:
//  1. Builds a local []registeredName from kept (only calls with a non-empty
//     literal name are included).
//  2. Exports that slice as a registeredNames package fact.
//  3. Scans AllPackageFacts() for every other package that has also exported a
//     registeredNames fact and collects name collisions.
//  4. Returns one Diagnostic per local site whose name collides, listing the
//     colliding package paths in sorted order.
func checkUniqueness(pass *analysis.Pass, kept []rules.MetricCall) []rules.Diagnostic {
	// Step 1 — build the local slice (literal names only).
	var local []registeredName
	for _, call := range kept {
		if call.Name == "" {
			continue // non-literal; string_literal rule will fire
		}
		local = append(local, registeredName{
			Name: call.Name,
			Pos:  pass.Fset.Position(call.Pos).String(),
		})
	}

	// Step 2 — export package fact (even if local is empty, so dependents can
	// scan us).
	fact := &registeredNames{Names: local}
	pass.ExportPackageFact(fact)

	if len(local) == 0 {
		return nil
	}

	// Step 3 — scan all other packages for name collisions.
	// Build map: name → []pkg path (from foreign packages only).
	nameToForeignPkgs := make(map[string][]string)
	for _, pf := range pass.AllPackageFacts() {
		rn, ok := pf.Fact.(*registeredNames)
		if !ok {
			continue
		}
		if pf.Package == pass.Pkg {
			continue // skip our own fact
		}
		for _, n := range rn.Names {
			nameToForeignPkgs[n.Name] = append(nameToForeignPkgs[n.Name], pf.Package.Path())
		}
	}

	// Sort the per-name package lists for deterministic output.
	for name := range nameToForeignPkgs {
		sort.Strings(nameToForeignPkgs[name])
	}

	// Step 4 — emit one diagnostic per local site with a collision.
	var diags []rules.Diagnostic
	for _, call := range kept {
		if call.Name == "" {
			continue
		}
		pkgs, collides := nameToForeignPkgs[call.Name]
		if !collides {
			continue
		}
		diags = append(diags, rules.Diagnostic{
			Pos:    call.Pos,
			RuleID: "cross_package_uniqueness",
			Message: fmt.Sprintf(
				"metric name %q also registered in: %s",
				call.Name,
				strings.Join(pkgs, ", "),
			),
		})
	}
	return diags
}
