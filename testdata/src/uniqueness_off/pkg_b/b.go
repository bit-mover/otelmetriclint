package pkg_b

import (
	"go.opentelemetry.io/otel/metric"

	_ "otelmetriclint_testdata/src/uniqueness_off/pkg_a"
)

func Register(m metric.Meter) {
	_, _ = m.Int64Counter("foo.bar")
}
