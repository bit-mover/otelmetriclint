package recognizer_enclosing

import "go.opentelemetry.io/otel/metric"

func outer(m metric.Meter) {
	func() {
		_, _ = m.Int64Counter("x") // want `enclosing=\[FuncDecl,FuncLit\]`
	}()
}

func topLevel(m metric.Meter) {
	_, _ = m.Int64Counter("y") // want `enclosing=\[FuncDecl\]`
}
