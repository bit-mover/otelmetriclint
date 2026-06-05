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
	case "Int64ObservableCounter":
		return rules.KindInt64ObservableCounter, true
	case "Float64ObservableCounter":
		return rules.KindFloat64ObservableCounter, true
	case "Int64ObservableUpDownCounter":
		return rules.KindInt64ObservableUpDownCounter, true
	case "Float64ObservableUpDownCounter":
		return rules.KindFloat64ObservableUpDownCounter, true
	case "Int64ObservableGauge":
		return rules.KindInt64ObservableGauge, true
	case "Float64ObservableGauge":
		return rules.KindFloat64ObservableGauge, true
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
//
// Calls inside helper-wrapper bodies are skipped: if the enclosing
// function or closure returns a metric instrument interface, the real
// metric-creation site is at the wrapper's callers, not in its body
// (where the name argument is necessarily a parameter, not a literal).
func findMetricCalls(pass *analysis.Pass, overrides []HelperOverride) []rules.MetricCall {
	var out []rules.MetricCall
	overrideIdx := buildOverrideIndex(overrides)
	for _, file := range pass.Files {
		walkWithFuncStack(file, nil, func(call *ast.CallExpr, funcStack []ast.Node) {
			if isInsideMetricHelper(pass, funcStack) {
				return
			}
			if mc, ok := recognizeCall(pass, call, overrideIdx); ok {
				mc.EnclosingFuncs = append([]ast.Node(nil), funcStack...)
				out = append(out, mc)
			}
		})
	}
	return out
}

// walkWithFuncStack walks the AST rooted at n in pre-order, maintaining
// a stack of enclosing *ast.FuncDecl or *ast.FuncLit nodes (one per
// enclosing function or closure). The visit callback is invoked for every
// *ast.CallExpr with the current stack snapshot. Function bodies are
// descended into recursively so the stack tracks the correct lexical scope.
func walkWithFuncStack(
	n ast.Node,
	stack []ast.Node,
	visit func(*ast.CallExpr, []ast.Node),
) {
	switch node := n.(type) {
	case nil:
		return
	case *ast.FuncDecl:
		// Push this FuncDecl; recurse into the body only.
		if node.Body != nil {
			walkWithFuncStack(node.Body, append(stack, ast.Node(node)), visit) //nolint:gocritic // intentional: a new slice keeps the caller's stack scoped
		}
		return
	case *ast.FuncLit:
		walkWithFuncStack(node.Body, append(stack, ast.Node(node)), visit) //nolint:gocritic // intentional: a new slice keeps the caller's stack scoped
		return
	case *ast.CallExpr:
		visit(node, stack)
		// Fall through so we still descend into the call's args
		// (which may themselves contain nested calls / func lits).
	}
	// Generic descent: walk every child of n with the current stack.
	for _, child := range childrenOf(n) {
		walkWithFuncStack(child, stack, visit)
	}
}

// childrenOf returns the direct AST children of n. Uses ast.Inspect with
// a depth counter — accepting only depth-1 descendants — to avoid having
// to enumerate every node type by hand.
func childrenOf(n ast.Node) []ast.Node {
	if n == nil {
		return nil
	}
	var kids []ast.Node
	ast.Inspect(n, func(c ast.Node) bool {
		if c == nil || c == n {
			return true
		}
		kids = append(kids, c)
		return false // do not descend further; walkWithFuncStack will recurse
	})
	return kids
}

// isInsideMetricHelper reports whether the current call site is inside
// a function or closure that returns a metric instrument interface.
// Such functions are wrappers (`MustInt64Counter`, application-level
// `counter := func(name, desc string) metric.Int64Counter { ... }`) —
// the real metric-creation site is at the wrapper's callers.
func isInsideMetricHelper(pass *analysis.Pass, funcStack []ast.Node) bool {
	for _, fn := range funcStack {
		if funcReturnsMetricInstrument(pass, fn) {
			return true
		}
	}
	return false
}

// funcReturnsMetricInstrument reports whether any result of fn is a
// metric instrument interface from go.opentelemetry.io/otel/metric.
// fn must be *ast.FuncDecl or *ast.FuncLit; other nodes return false.
func funcReturnsMetricInstrument(pass *analysis.Pass, fn ast.Node) bool {
	var ft *ast.FuncType
	switch n := fn.(type) {
	case *ast.FuncDecl:
		ft = n.Type
	case *ast.FuncLit:
		ft = n.Type
	default:
		return false
	}
	if ft == nil || ft.Results == nil {
		return false
	}
	for _, result := range ft.Results.List {
		t := pass.TypesInfo.TypeOf(result.Type)
		if t == nil {
			continue
		}
		if _, ok := matchInstrumentKind(t); ok {
			return true
		}
	}
	return false
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

// stringLiteralValue returns the resolved string value of e when e is a
// compile-time constant string expression. This includes:
//   - string literals ("foo.bar")
//   - const identifiers (const name = "foo.bar")
//   - string conversions of const expressions (string(constExpr))
//   - constant concatenation ("foo." + "bar", constA + constB)
//   - imported package-level string constants (pkg.ConstName)
//
// Non-constant expressions (runtime variables, function calls, fmt.Sprintf,
// etc.) return ("", false) so the string_literal rule can report them.
func stringLiteralValue(pass *analysis.Pass, e ast.Expr) (string, bool) {
	tv, ok := pass.TypesInfo.Types[e]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return "", false
	}
	return constant.StringVal(tv.Value), true
}
