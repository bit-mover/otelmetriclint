package bad_observable_total

import "go.opentelemetry.io/otel/metric"

func observableCounterTotal(m metric.Meter) {
	// Observable counter ending in _total fires total_suffix (monotonic counter)
	// and unit_suffix ("total" is a forbidden unit-like suffix).
	_, _ = m.Int64ObservableCounter("eventhub.drops_total") // want `total_suffix: counter name "eventhub.drops_total" must not end in _total` `unit_suffix: metric name "eventhub.drops_total" has unit-like suffix "total"`
}
