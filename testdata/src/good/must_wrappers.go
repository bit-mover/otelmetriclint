package good

import (
	"go.opentelemetry.io/otel/metric"
)

func mustInt64Counter(m metric.Meter, name string, opts ...metric.Int64CounterOption) metric.Int64Counter {
	c, err := m.Int64Counter(name, opts...) // want `found Int64Counter name=""`
	if err != nil {
		panic(err)
	}
	return c
}

func mustFloat64Histogram(m metric.Meter, name string, opts ...metric.Float64HistogramOption) metric.Float64Histogram {
	h, err := m.Float64Histogram(name, opts...) // want `found Float64Histogram name=""`
	if err != nil {
		panic(err)
	}
	return h
}

func wrapperUsage(m metric.Meter) {
	_ = mustInt64Counter(m, "wrapped.counter")                             // want `found Int64Counter name="wrapped.counter"`
	_ = mustFloat64Histogram(m, "wrapped.histogram", metric.WithUnit("s")) // want `found Float64Histogram name="wrapped.histogram"`
}
