package bad_ucum_unit

import "go.opentelemetry.io/otel/metric"

func typo(m metric.Meter) {
	_, _ = m.Float64Histogram("worker.poll.interval", metric.WithUnit("secnds")) // want `ucum_unit: metric unit "secnds" is not a valid UCUM code`
}
