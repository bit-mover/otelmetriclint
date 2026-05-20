package bad_unit_suffix

import "go.opentelemetry.io/otel/metric"

func ld(m metric.Meter) {
	_, _ = m.Float64Histogram("foo.load_duration", metric.WithUnit("s")) // want `unit_suffix: metric name "foo.load_duration" has unit-like suffix "duration"`
}
