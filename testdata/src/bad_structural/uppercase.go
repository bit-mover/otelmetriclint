package bad_structural

import "go.opentelemetry.io/otel/metric"

func uppercase(m metric.Meter) {
	_, _ = m.Int64Counter("Foo.bar") // want `structural: metric name "Foo.bar" must be ASCII-lowercase`
}
