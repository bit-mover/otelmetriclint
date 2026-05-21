//nolint:otelmetriclint
package suppress_file_level

import "go.opentelemetry.io/otel/metric"

func f(m metric.Meter) {
	_, _ = m.Int64Counter("anything_total")
	_, _ = m.Int64Counter("everything_total")
}
