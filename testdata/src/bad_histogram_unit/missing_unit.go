package bad_histogram_unit

import "go.opentelemetry.io/otel/metric"

func noUnit(m metric.Meter) {
	_, _ = m.Float64Histogram("foo.bar") // want `histogram_unit: histogram "foo.bar" must set metric.WithUnit`
}
