package bad_const_name

import "go.opentelemetry.io/otel/metric"

// counterName is a package-level const whose resolved value ends in _total.
// The recognizer must resolve it to the string "eventhub.drops_total" so that
// the total_suffix rule (and unit_suffix rule) fire — identical to what they
// would do for a literal string argument.
const counterName = "eventhub.drops_total"

func constTotalCounter(m metric.Meter) {
	_, _ = m.Int64Counter(counterName) // want `total_suffix: counter name "eventhub.drops_total" must not end in _total` `unit_suffix: metric name "eventhub.drops_total" has unit-like suffix "total"`
}
