package good

import (
	"go.opentelemetry.io/otel/metric"
)

func sdkMethods(m metric.Meter) {
	_, _ = m.Int64Counter("good.counter")                                   // want `found Int64Counter name="good.counter"`
	_, _ = m.Float64Counter("good.float_counter")                           // want `found Float64Counter name="good.float_counter"`
	_, _ = m.Int64UpDownCounter("good.updown")                              // want `found Int64UpDownCounter name="good.updown"`
	_, _ = m.Float64UpDownCounter("good.float_updown")                      // want `found Float64UpDownCounter name="good.float_updown"`
	_, _ = m.Int64Histogram("good.histogram", metric.WithUnit("s"))         // want `found Int64Histogram name="good.histogram"`
	_, _ = m.Float64Histogram("good.float_histogram", metric.WithUnit("s")) // want `found Float64Histogram name="good.float_histogram"`
	_, _ = m.Int64Gauge("good.gauge")                                       // want `found Int64Gauge name="good.gauge"`
	_, _ = m.Float64Gauge("good.float_gauge")                               // want `found Float64Gauge name="good.float_gauge"`
}
