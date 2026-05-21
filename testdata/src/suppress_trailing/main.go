package suppress_trailing

import "go.opentelemetry.io/otel/metric"

func f(m metric.Meter) {
	_, _ = m.Int64Counter("good_total") //nolint:otelmetriclint

	_, _ = m.Int64Counter("bad_total") // want `total_suffix: counter name "bad_total" must not end in _total` `unit_suffix: metric name "bad_total" has unit-like suffix "total"`
}
