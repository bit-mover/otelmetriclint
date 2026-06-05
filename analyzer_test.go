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
