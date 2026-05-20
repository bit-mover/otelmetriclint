package otelmetriclint

import (
	"go/ast"
	"go/constant"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/bit-mover/otelmetriclint/rules"
)

const otelMetricPkgPath = "go.opentelemetry.io/otel/metric"

// matchInstrumentKind reports the rules.InstrumentKind corresponding to t
// if t is one of the metric instrument interfaces from
// go.opentelemetry.io/otel/metric. Returns (KindUnknown, false) otherwise.
func matchInstrumentKind(t types.Type) (rules.InstrumentKind, bool) {
	named, ok := t.(*types.Named)
	if !ok {
		return rules.KindUnknown, false
	}
	obj := named.Obj()
	if obj.Pkg() == nil || obj.Pkg().Path() != otelMetricPkgPath {
		return rules.KindUnknown, false
	}
	switch obj.Name() {
	case "Int64Counter":
		return rules.KindInt64Counter, true
	case "Float64Counter":
		return rules.KindFloat64Counter, true
	case "Int64UpDownCounter":
		return rules.KindInt64UpDownCounter, true
	case "Float64UpDownCounter":
		return rules.KindFloat64UpDownCounter, true
	case "Int64Histogram":
		return rules.KindInt64Histogram, true
	case "Float64Histogram":
		return rules.KindFloat64Histogram, true
	case "Int64Gauge":
		return rules.KindInt64Gauge, true
	case "Float64Gauge":
		return rules.KindFloat64Gauge, true
	}
	return rules.KindUnknown, false
}

// HelperOverride lets users declare wrappers where the metric name is not
// the first string-literal argument. Type-driven recognition handles the
// common case (SDK methods and Must-style wrappers) for free; overrides
// are for unusual wrapper signatures.
type HelperOverride struct {
	Pkg     string
	Func    string
	NameArg int // zero-indexed position of the name argument
}

// findMetricCalls walks pass.Files and returns every recognized
// MetricCall. overrides is optional; nil means "use type-driven defaults".
func findMetricCalls(pass *analysis.Pass, overrides []HelperOverride) []rules.MetricCall {
	var out []rules.MetricCall
	overrideIdx := buildOverrideIndex(overrides)
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			mc, ok := recognizeCall(pass, call, overrideIdx)
			if ok {
				out = append(out, mc)
			}
			return true
		})
	}
	return out
}

type overrideKey struct {
	Pkg  string
	Func string
}

func buildOverrideIndex(overrides []HelperOverride) map[overrideKey]int {
	if len(overrides) == 0 {
		return nil
	}
	idx := make(map[overrideKey]int, len(overrides))
	for _, o := range overrides {
		idx[overrideKey{Pkg: o.Pkg, Func: o.Func}] = o.NameArg
	}
	return idx
}

func recognizeCall(pass *analysis.Pass, call *ast.CallExpr, overrides map[overrideKey]int) (rules.MetricCall, bool) {
	typ := pass.TypesInfo.TypeOf(call)
	if typ == nil {
		return rules.MetricCall{}, false
	}
	kind := instrumentKindFromCallType(typ)
	if kind == rules.KindUnknown {
		return rules.MetricCall{}, false
	}
	nameArgIdx := resolveNameArgIndex(pass, call, overrides)
	if nameArgIdx < 0 || nameArgIdx >= len(call.Args) {
		return rules.MetricCall{}, false
	}
	nameArg := call.Args[nameArgIdx]
	name, _ := stringLiteralValue(pass, nameArg)
	// name == "" when the arg isn't a string literal — recognizer still
	// emits the MetricCall so the string_literal rule (Task 3.2) reports it.
	return rules.MetricCall{
		Pos:     call.Pos(),
		Kind:    kind,
		Name:    name,
		NameArg: nameArg,
		Options: call.Args[nameArgIdx+1:],
	}, true
}

// instrumentKindFromCallType handles both single returns (wrappers) and
// tuple returns (SDK factory methods returning (Instrument, error)).
func instrumentKindFromCallType(t types.Type) rules.InstrumentKind {
	if tt, ok := t.(*types.Tuple); ok {
		if tt.Len() != 2 {
			return rules.KindUnknown
		}
		kind, _ := matchInstrumentKind(tt.At(0).Type())
		return kind
	}
	kind, _ := matchInstrumentKind(t)
	return kind
}

// resolveNameArgIndex returns the position of the metric-name argument.
// With no overrides, the heuristic is "first string-typed argument"; for
// SDK methods this is arg 0, for Must-style wrappers it is arg 1.
func resolveNameArgIndex(pass *analysis.Pass, call *ast.CallExpr, overrides map[overrideKey]int) int {
	if len(overrides) > 0 {
		pkg, fn := callTarget(pass, call)
		if pkg != "" && fn != "" {
			if idx, ok := overrides[overrideKey{Pkg: pkg, Func: fn}]; ok {
				return idx
			}
		}
	}
	return firstStringArgIndex(pass, call)
}

// firstStringArgIndex returns the index of the first call argument whose
// type is `string` (or untyped string). Returns -1 if none.
func firstStringArgIndex(pass *analysis.Pass, call *ast.CallExpr) int {
	for i, arg := range call.Args {
		if isStringTyped(pass, arg) {
			return i
		}
	}
	return -1
}

func isStringTyped(pass *analysis.Pass, e ast.Expr) bool {
	tv, ok := pass.TypesInfo.Types[e]
	if !ok || tv.Type == nil {
		return false
	}
	if basic, ok := tv.Type.Underlying().(*types.Basic); ok {
		return basic.Kind() == types.String || basic.Kind() == types.UntypedString
	}
	return false
}

// callTarget returns the package path and function/method name of the
// function being called, when resolvable. Used for matching overrides.
func callTarget(pass *analysis.Pass, call *ast.CallExpr) (pkg, fn string) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", ""
	}
	if selInfo := pass.TypesInfo.Selections[sel]; selInfo != nil {
		recv := selInfo.Recv()
		if named, ok := recv.(*types.Named); ok && named.Obj().Pkg() != nil {
			return named.Obj().Pkg().Path(), sel.Sel.Name
		}
	}
	if id, ok := sel.X.(*ast.Ident); ok {
		if obj := pass.TypesInfo.ObjectOf(id); obj != nil {
			if pkgName, ok := obj.(*types.PkgName); ok {
				return pkgName.Imported().Path(), sel.Sel.Name
			}
		}
	}
	return "", ""
}

// stringLiteralValue returns the unquoted string value of e if it is a
// *ast.BasicLit of kind STRING. Returns ("", false) for non-literal
// expressions (identifiers, calls, etc.) so the string_literal rule can fire.
func stringLiteralValue(pass *analysis.Pass, e ast.Expr) (string, bool) {
	tv, ok := pass.TypesInfo.Types[e]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return "", false
	}
	if _, isLit := e.(*ast.BasicLit); !isLit {
		return "", false
	}
	return constant.StringVal(tv.Value), true
}
