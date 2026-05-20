package good_prefix

import "go.opentelemetry.io/otel/metric"

func allowedUse(m metric.Meter) {
	_, _ = m.Int64Counter("allowed.thing")
}
