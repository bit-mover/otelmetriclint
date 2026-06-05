package pkg_b // want package:`registeredNames\{foo\.bar@.*b\.go:10:9\}`

import (
	"go.opentelemetry.io/otel/metric"

	_ "otelmetriclint_testdata/src/uniqueness/pkg_a"
)

func Register(m metric.Meter) {
	_, _ = m.Int64Counter("foo.bar") // want `cross_package_uniqueness: metric name "foo\.bar" also registered in: otelmetriclint_testdata/src/uniqueness/pkg_a`
}
