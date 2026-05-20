package bad_unit_suffix

import "go.opentelemetry.io/otel/metric"

func by(m metric.Meter) {
	_, _ = m.Int64Counter("foo.bytes") // want `unit_suffix: metric name "foo.bytes" has unit-like suffix "bytes"`
}
