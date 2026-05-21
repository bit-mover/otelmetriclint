package suppress_above_funclit

import "go.opentelemetry.io/otel/metric"

func Outer(m metric.Meter) {
	//nolint:otelmetriclint
	fn := func() {
		_, _ = m.Int64Counter("closure_total")
	}
	fn()

	other := func() {
		_, _ = m.Int64Counter("noisy_total") // want `total_suffix: counter name "noisy_total" must not end in _total` `unit_suffix: metric name "noisy_total" has unit-like suffix "total"`
	}
	other()
}
