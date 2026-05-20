package bad_unit_suffix

import "go.opentelemetry.io/otel/metric"

func dur(m metric.Meter) {
	_, _ = m.Int64Counter("foo.duration") // want `unit_suffix: metric name "foo.duration" has unit-like suffix "duration"`
}
