package bad_observable_gauge

import "go.opentelemetry.io/otel/metric"

// Int64ObservableGauge is non-monotonic and is NOT a counter, so _total suffix
// must NOT fire total_suffix. However, unit_suffix still fires because "total"
// is a forbidden unit-like suffix for all instrument kinds.
func observableGaugeTotalSuffix(m metric.Meter) {
	_, _ = m.Int64ObservableGauge("foo.bar_total") // want `unit_suffix: metric name "foo.bar_total" has unit-like suffix "total"`
}
