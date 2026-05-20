package otelmetriclint

import (
	"golang.org/x/tools/go/analysis"

	"github.com/bit-mover/otelmetriclint/rules"
)

// New returns an Analyzer configured by cfg. Pass Default() for the
// built-in defaults.
func New(cfg Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "otelmetriclint",
		Doc:  "validates OpenTelemetry metric instrument creation call sites",
		URL:  "https://github.com/bit-mover/otelmetriclint",
		Run:  makeRun(cfg),
	}
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

	return func(pass *analysis.Pass) (interface{}, error) {
		calls := findMetricCalls(pass, overrides)
		for _, mc := range calls {
			for _, rule := range enabled {
				for _, d := range rule.Check(mc) {
					pass.Reportf(d.Pos, "%s: %s", d.RuleID, d.Message)
				}
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
