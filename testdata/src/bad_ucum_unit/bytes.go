package bad_ucum_unit

import "go.opentelemetry.io/otel/metric"

// The calls in validUnits use valid UCUM codes. No diagnostic should fire on them.
func validUnits(m metric.Meter) {
	_, _ = m.Int64Counter("cache.read.op", metric.WithUnit("s"))
	_, _ = m.Int64Counter("cache.write.op", metric.WithUnit("By"))
	_, _ = m.Float64Histogram("network.transfer.rate", metric.WithUnit("By/s"))
	_, _ = m.Int64Counter("system.event.op", metric.WithUnit("{event}"))
	_, _ = m.Float64Gauge("sensor.area.reading", metric.WithUnit("m2"))
}

func by(m metric.Meter) {
	_, _ = m.Int64Counter("cache.read.size", metric.WithUnit("bytes")) // want `ucum_unit: metric unit "bytes" is not a valid UCUM code`
}
