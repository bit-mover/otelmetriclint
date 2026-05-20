package bad_unit_suffix

import "go.opentelemetry.io/otel/metric"

func sec(m metric.Meter) {
	_, _ = m.Int64Counter("foo.seconds") // want `unit_suffix: metric name "foo.seconds" has unit-like suffix "seconds"`
}
