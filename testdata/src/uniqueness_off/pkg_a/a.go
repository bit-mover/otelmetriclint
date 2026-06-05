package pkg_a

import "go.opentelemetry.io/otel/metric"

func Register(m metric.Meter) {
	_, _ = m.Int64Counter("foo.bar")
}
