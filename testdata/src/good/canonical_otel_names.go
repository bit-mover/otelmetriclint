package good

import "go.opentelemetry.io/otel/metric"

// Canonical OTel semconv metric names that use quantity descriptors
// (`duration`, `count`) — these must NOT be diagnosed by the unit_suffix
// rule because the descriptors aren't units (the unit goes in WithUnit).
// See https://github.com/open-telemetry/semantic-conventions for the
// canonical examples below.

func canonicalDurationHistograms(m metric.Meter) {
	_, _ = m.Float64Histogram("http.server.request.duration", metric.WithUnit("s"))    // want `found Float64Histogram name="http.server.request.duration"`
	_, _ = m.Float64Histogram("http.client.request.duration", metric.WithUnit("s"))    // want `found Float64Histogram name="http.client.request.duration"`
	_, _ = m.Float64Histogram("db.client.operation.duration", metric.WithUnit("s"))    // want `found Float64Histogram name="db.client.operation.duration"`
	_, _ = m.Float64Histogram("foo.load_duration", metric.WithUnit("s"))               // want `found Float64Histogram name="foo.load_duration"`
}

func canonicalCountCounters(m metric.Meter) {
	_, _ = m.Int64UpDownCounter("db.client.connection.count") // want `found Int64UpDownCounter name="db.client.connection.count"`
	_, _ = m.Int64UpDownCounter("process.thread.count")       // want `found Int64UpDownCounter name="process.thread.count"`
}
