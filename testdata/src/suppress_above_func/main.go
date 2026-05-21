package suppress_above_func

import "go.opentelemetry.io/otel/metric"

//nolint:otelmetriclint
func WithFreeFloatingComment(m metric.Meter) {
	_, _ = m.Int64Counter("free_total")
	_, _ = m.Int64Counter("anything_total")
}

// WithDocComment carries the directive inside its doc comment.
//
//nolint:otelmetriclint
func WithDocComment(m metric.Meter) {
	_, _ = m.Int64Counter("doc_total")
}

func NotSuppressed(m metric.Meter) {
	_, _ = m.Int64Counter("loud_total") // want `total_suffix: counter name "loud_total" must not end in _total` `unit_suffix: metric name "loud_total" has unit-like suffix "total"`
}
