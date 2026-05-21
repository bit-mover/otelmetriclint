package otelmetriclint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/analysis"

	"github.com/bit-mover/otelmetriclint/rules"
)

func TestMatchesNoLint(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		// Positive — bare nolint suppresses everything.
		{"//nolint", true},
		{"//nolint ", true},
		{"//nolint\t", true},

		// Positive — linter list including otelmetriclint.
		{"//nolint:otelmetriclint", true},
		{"//nolint:otelmetriclint,gocritic", true},
		{"//nolint:gocritic,otelmetriclint", true},
		{"//nolint:gocritic, otelmetriclint", true},
		{"//nolint:otelmetriclint ", true},
		{"//nolint:otelmetriclint // legacy migration", true},
		{"/*nolint:otelmetriclint*/", true},
		{"/* nolint:otelmetriclint */", true},

		// Negative — wrong shape.
		{"// nolint:otelmetriclint", false}, // golangci's space-after-// quirk
		{"//NOLINT:otelmetriclint", false},  // wrong case
		{"//nolint:gocritic", false},
		{"//nolint:otelmetriclintfoo", false}, // not exact-name
		{"//notanolint:otelmetriclint", false},
		{"//nolintfoo", false},
		{"// just a comment", false},
		{"", false},
		// "//nolint:" with an empty list is NOT treated as bare //nolint;
		// it matches nothing. Mirrors golangci-lint.
		{"//nolint:", false},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			if got := matchesNoLint(c.in); got != c.want {
				t.Errorf("matchesNoLint(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

// passFromSource parses src as Go source and returns a minimal *analysis.Pass
// suitable for buildSuppressIndex. It does NOT run type-checking; the
// suppression layer only needs Fset, Files, and the comment groups.
func passFromSource(t *testing.T, src string) *analysis.Pass {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "src.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return &analysis.Pass{
		Fset:      fset,
		Files:     []*ast.File{file},
		TypesInfo: &types.Info{},
	}
}

// callAt returns a synthetic MetricCall whose Pos lands on the given line
// of pass.Files[0]. EnclosingFuncs and other fields are left zero.
func callAt(t *testing.T, pass *analysis.Pass, line int) rules.MetricCall {
	t.Helper()
	tokFile := pass.Fset.File(pass.Files[0].Pos())
	return rules.MetricCall{Pos: tokFile.LineStart(line)}
}

func TestSuppressed_Trailing(t *testing.T) {
	src := `package p

func f() {
    _ = "suppressed" //nolint:otelmetriclint
    _ = "not suppressed"
}
`
	pass := passFromSource(t, src)
	idx := buildSuppressIndex(pass)

	suppressedLine := 4
	notSuppressedLine := 5

	if !idx.suppressed(callAt(t, pass, suppressedLine)) {
		t.Errorf("call on line %d: want suppressed=true, got false", suppressedLine)
	}
	if idx.suppressed(callAt(t, pass, notSuppressedLine)) {
		t.Errorf("call on line %d: want suppressed=false, got true", notSuppressedLine)
	}
}
