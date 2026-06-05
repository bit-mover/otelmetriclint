package pkg_a // want package:`registeredNames\{foo\.bar@.*a\.go:6:9\}`

import "go.opentelemetry.io/otel/metric"

func Register(m metric.Meter) {
	_, _ = m.Int64Counter("foo.bar")
}
