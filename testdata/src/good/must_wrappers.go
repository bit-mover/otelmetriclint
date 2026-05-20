package good

import (
	"go.opentelemetry.io/otel/metric"
)

// The inner SDK calls in these wrappers should NOT produce diagnostics —
// the enclosing function returns a metric instrument, marking it as a
// helper wrapper. The real metric-creation site is at the wrapper's
// callers (see wrapperUsage below), where the name is a literal.
func mustInt64Counter(m metric.Meter, name string, opts ...metric.Int64CounterOption) metric.Int64Counter {
	c, err := m.Int64Counter(name, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

func mustFloat64Histogram(m metric.Meter, name string, opts ...metric.Float64HistogramOption) metric.Float64Histogram {
	h, err := m.Float64Histogram(name, opts...)
	if err != nil {
		panic(err)
	}
	return h
}

func wrapperUsage(m metric.Meter) {
	_ = mustInt64Counter(m, "wrapped.counter")                             // want `found Int64Counter name="wrapped.counter"`
	_ = mustFloat64Histogram(m, "wrapped.histogram", metric.WithUnit("s")) // want `found Float64Histogram name="wrapped.histogram"`
}

// Closure helpers — same suppression should apply: the closure returns
// a metric instrument, so the call inside its body is not the real
// metric-creation site. Callers that pass literals ARE the real sites.
func closureWrapperUsage(m metric.Meter) {
	counter := func(name, desc string) metric.Int64Counter {
		return mustInt64Counter(m, name, metric.WithDescription(desc))
	}
	_ = counter("closure.counter", "via closure") // want `found Int64Counter name="closure.counter"`
}
