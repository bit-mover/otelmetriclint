package otelmetriclint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
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

// lineOf returns the 1-indexed line of the first occurrence of marker in
// src. Use to anchor test assertions against the source text rather than
// hard-coded line numbers, which silently break when the source string
// is edited.
func lineOf(t *testing.T, src, marker string) int {
	t.Helper()
	idx := strings.Index(src, marker)
	if idx < 0 {
		t.Fatalf("marker %q not found in source", marker)
	}
	return strings.Count(src[:idx], "\n") + 1
}

func TestSuppressed_AboveCall(t *testing.T) {
	src := `package p

func f() {
    //nolint:otelmetriclint
    _ = "suppressed"

    //nolint:otelmetriclint

    _ = "blank-line gap, not suppressed"
}
`
	pass := passFromSource(t, src)
	idx := buildSuppressIndex(pass)

	suppressedLine := lineOf(t, src, `"suppressed"`)
	notSuppressedLine := lineOf(t, src, `"blank-line gap, not suppressed"`)

	if !idx.suppressed(callAt(t, pass, suppressedLine)) {
		t.Errorf("line %d (directive immediately above): want suppressed=true", suppressedLine)
	}
	if idx.suppressed(callAt(t, pass, notSuppressedLine)) {
		t.Errorf("line %d (blank line between directive and call): want suppressed=false", notSuppressedLine)
	}
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

	suppressedLine := lineOf(t, src, `"suppressed"`)
	notSuppressedLine := lineOf(t, src, `"not suppressed"`)

	if !idx.suppressed(callAt(t, pass, suppressedLine)) {
		t.Errorf("call on line %d: want suppressed=true, got false", suppressedLine)
	}
	if idx.suppressed(callAt(t, pass, notSuppressedLine)) {
		t.Errorf("call on line %d: want suppressed=false, got true", notSuppressedLine)
	}
}

func TestSuppressed_AboveFunc(t *testing.T) {
	src := `package p

//nolint:otelmetriclint
func WithFreeFloatingComment() {
    _ = "suppressed (free-floating above func)"
}

// WithDocComment does X.
//
//nolint:otelmetriclint
func WithDocComment() {
    _ = "suppressed (doc comment carries directive)"
}

func NotSuppressed() {
    _ = "not suppressed"
}

func WithClosure() {
    //nolint:otelmetriclint
    fn := func() {
        _ = "suppressed (above closure)"
    }
    _ = fn
}
`
	pass := passFromSource(t, src)
	idx := buildSuppressIndex(pass)

	file := pass.Files[0]
	funcByName := map[string]ast.Node{}
	var funcLit *ast.FuncLit
	ast.Inspect(file, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok {
			funcByName[fd.Name.Name] = fd
		}
		if fl, ok := n.(*ast.FuncLit); ok {
			funcLit = fl
		}
		return true
	})

	tests := []struct {
		name           string
		marker         string
		enclosing      []ast.Node
		wantSuppressed bool
	}{
		{"free-floating above func", `"suppressed (free-floating above func)"`, []ast.Node{funcByName["WithFreeFloatingComment"]}, true},
		{"doc comment carries directive", `"suppressed (doc comment carries directive)"`, []ast.Node{funcByName["WithDocComment"]}, true},
		{"no directive on this func", `"not suppressed"`, []ast.Node{funcByName["NotSuppressed"]}, false},
		{"above closure", `"suppressed (above closure)"`, []ast.Node{funcByName["WithClosure"], funcLit}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := lineOf(t, src, tt.marker)
			call := callAt(t, pass, line)
			call.EnclosingFuncs = tt.enclosing
			if got := idx.suppressed(call); got != tt.wantSuppressed {
				t.Errorf("line %d: got suppressed=%v, want %v", line, got, tt.wantSuppressed)
			}
		})
	}
}
