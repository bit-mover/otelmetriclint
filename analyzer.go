package otelmetriclint

import (
	"golang.org/x/tools/go/analysis"

	"github.com/bit-mover/otelmetriclint/rules"
)

// New returns an Analyzer configured by cfg. Pass Default() for the
// built-in defaults.
func New(cfg Config) *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name: "otelmetriclint",
		Doc:  "validates OpenTelemetry metric instrument creation call sites",
		URL:  "https://github.com/bit-mover/otelmetriclint",
		Run:  makeRun(cfg),
	}
	// Register the package fact only when the rule is enabled, so that
	// packages analyzed without the rule never pay the gob-encoding cost.
	if enabled, ok := cfg.Rules["cross_package_uniqueness"]; ok && enabled {
		a.FactTypes = []analysis.Fact{new(registeredNames)}
	}
	return a
}

// DefaultAnalyzer returns an Analyzer with built-in defaults.
func DefaultAnalyzer() *analysis.Analyzer {
	return New(Default())
}

func makeRun(cfg Config) func(*analysis.Pass) (interface{}, error) {
	reg := rules.Registry{Rules: []rules.Rule{
		rules.StringLiteral{},
		rules.Structural{},
		rules.Prefix{Allowed: cfg.Prefix.Allowed},
		rules.TotalSuffix{},
		rules.UnitSuffix{Forbidden: mergeUnitSuffix(cfg.UnitSuffix)},
		rules.HistogramUnit{},
	}}
	overrides := make([]HelperOverride, len(cfg.Helpers))
	for i, h := range cfg.Helpers {
		overrides[i] = HelperOverride(h)
	}
	enabled := reg.Enabled(cfg.Rules)
	uniquenessEnabled := cfg.Rules["cross_package_uniqueness"]

	return func(pass *analysis.Pass) (interface{}, error) {
		calls := findMetricCalls(pass, overrides)
		sup := buildSuppressIndex(pass)

		// Compute the non-suppressed calls once; reused by both the rule loop
		// and checkUniqueness so we never traverse twice.
		kept := make([]rules.MetricCall, 0, len(calls))
		for _, mc := range calls {
			if !sup.suppressed(mc) {
				kept = append(kept, mc)
			}
		}

		for _, mc := range kept {
			for _, rule := range enabled {
				for _, d := range rule.Check(mc) {
					pass.Reportf(d.Pos, "%s: %s", d.RuleID, d.Message)
				}
			}
		}

		if uniquenessEnabled {
			for _, d := range checkUniqueness(pass, kept) {
				pass.Reportf(d.Pos, "%s: %s", d.RuleID, d.Message)
			}
		}

		return nil, nil
	}
}

// mergeUnitSuffix concatenates the forbidden list with any project-specific
// additions so the rule sees one effective list.
func mergeUnitSuffix(c UnitSuffixConfig) []string {
	out := make([]string, 0, len(c.Forbidden)+len(c.Additional))
	out = append(out, c.Forbidden...)
	out = append(out, c.Additional...)
	return out
}
