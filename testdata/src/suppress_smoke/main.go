package suppress_smoke

import "go.opentelemetry.io/otel/metric"

func f(m metric.Meter) {
	_, _ = m.Int64Counter("eventhub.drops_total") //nolint:otelmetriclint
}
