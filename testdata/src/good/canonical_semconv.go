package good

import "go.opentelemetry.io/otel/metric"

// Canonical registered OTel semconv metric names — these must NOT produce any
// semconv diagnostic even when the rule is enabled.

func canonicalSemconvNames(m metric.Meter) {
	_, _ = m.Int64Counter("http.server.active_requests")                           // want `found Int64Counter name="http\.server\.active_requests"`
	_, _ = m.Float64Histogram("http.server.request.body.size", metric.WithUnit("By")) // want `found Float64Histogram name="http\.server\.request\.body\.size"`
}
