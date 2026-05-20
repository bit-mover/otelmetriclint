package bad_total_suffix

import "go.opentelemetry.io/otel/metric"

func counterTotal(m metric.Meter) {
	// _total fires both rules: total_suffix (counter-specific) and unit_suffix ("total" is a forbidden unit-like suffix).
	_, _ = m.Int64Counter("eventhub.drops_total") // want `total_suffix: counter name "eventhub.drops_total" must not end in _total` `unit_suffix: metric name "eventhub.drops_total" has unit-like suffix "total"`
}
