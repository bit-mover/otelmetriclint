package bad_ucum_unit

import "go.opentelemetry.io/otel/metric"

func sec(m metric.Meter) {
	_, _ = m.Float64Histogram("task.queue.duration", metric.WithUnit("seconds")) // want `ucum_unit: metric unit "seconds" is not a valid UCUM code`
}
