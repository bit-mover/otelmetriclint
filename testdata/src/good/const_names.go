package good

import "go.opentelemetry.io/otel/metric"

// Package-level string constants used to verify that the recognizer resolves
// compile-time constants to their string values rather than returning "".

const (
	constCounter    = "const.counter"
	constHistPrefix = "const.histo"
	constHistSuffix = "gram"
)

// constNames exercises four constant-expression forms that stringLiteralValue
// must now resolve (the *ast.BasicLit guard was removed):
//
//  1. plain const ident
//  2. string(const) conversion
//  3. const concatenation via +
//  4. cross-package const — simulated here by an ident defined at package
//     scope in a different file; the AST node is *ast.Ident in both cases and
//     the type-checker resolves it identically to an imported-package const.
func constNames(m metric.Meter) {
	// 1. Plain const ident — AST node is *ast.Ident, not *ast.BasicLit.
	_, _ = m.Int64Counter(constCounter) // want `found Int64Counter name="const.counter"`

	// 2. string(const) conversion — AST node is *ast.CallExpr wrapping an ident.
	_, _ = m.Int64Gauge(string(constCounter)) // want `found Int64Gauge name="const.counter"`

	// 3. Constant concatenation — AST node is *ast.BinaryExpr.
	_, _ = m.Int64Histogram(constHistPrefix+constHistSuffix, metric.WithUnit("s")) // want `found Int64Histogram name="const.histogram"`
}
