package otelmetriclint

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzerDefault runs the real analyzer (default config) against the
// bad_* fixtures. Each fixture is a small Go file with `// want` annotations
// matching the rule diagnostics that should fire at that call site.
//
// bad_string_literal is intentionally excluded: its fixtures already carry
// `// want \"found ... name=\\\"\\\"\"` annotations targeting the recognizer's
// debug analyzer (see recognizer_test.go). The string_literal rule itself is
// covered by the direct table test in rules/string_literal_test.go.
func TestAnalyzerDefault(t *testing.T) {
	a := DefaultAnalyzer()
	for _, dir := range []string{
		"./src/bad_structural",
		"./src/bad_total_suffix",
		"./src/bad_unit_suffix",
		"./src/bad_histogram_unit",
		"./src/bad_const_name",
		"./src/bad_observable_total",
		"./src/bad_observable_gauge",
		"./src/bad_observable_structural",
		"./src/bad_pluralization",
		"./src/bad_ucum_unit",
		"./src/suppress_trailing",
		"./src/suppress_above_call",
		"./src/suppress_above_func",
		"./src/suppress_above_funclit",
		"./src/suppress_file_level",
		"./src/suppress_negative",
	} {
		t.Run(dir, func(t *testing.T) {
			analysistest.Run(t, analysistest.TestData(), a, dir)
		})
	}
}

// TestAnalyzerWithPrefixConfig exercises the prefix rule, which is off by
// default. The config loaded here enables it with a one-name allowlist.
func TestAnalyzerWithPrefixConfig(t *testing.T) {
	cfg, err := Load("testdata/configs/prefix_allowed.yaml")
	if err != nil {
		t.Fatal(err)
	}
	a := New(cfg)
	analysistest.Run(t, analysistest.TestData(), a, "./src/good_prefix")
	analysistest.Run(t, analysistest.TestData(), a, "./src/bad_prefix")
}

// TestAnalyzerCrossPackageUniqueness verifies that when cross_package_uniqueness
// is enabled, a metric name registered in both pkg_a and pkg_b produces a
// diagnostic at the pkg_b call site listing pkg_a's import path.
func TestAnalyzerCrossPackageUniqueness(t *testing.T) {
	cfg, err := Load("testdata/configs/uniqueness.yaml")
	if err != nil {
		t.Fatal(err)
	}
	a := New(cfg)
	analysistest.Run(t, analysistest.TestData(), a,
		"./src/uniqueness/pkg_a",
		"./src/uniqueness/pkg_b",
	)
}

// TestAnalyzerCrossPackageUniquenessOffByDefault verifies that with the default
// config (cross_package_uniqueness disabled) no diagnostic is emitted even when
// the same metric name appears in two packages. The uniqueness_off fixtures
// contain no // want annotations so analysistest would fail if any diagnostic fired.
func TestAnalyzerCrossPackageUniquenessOffByDefault(t *testing.T) {
	a := DefaultAnalyzer()
	analysistest.Run(t, analysistest.TestData(), a,
		"./src/uniqueness_off/pkg_a",
		"./src/uniqueness_off/pkg_b",
	)
}
