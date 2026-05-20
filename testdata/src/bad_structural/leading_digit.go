package bad_structural

import "go.opentelemetry.io/otel/metric"

func leading(m metric.Meter) {
	_, _ = m.Int64Counter("9foo") // want `structural: metric name "9foo" must be ASCII-lowercase`
}
