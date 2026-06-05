package semconv_off

import "go.opentelemetry.io/otel/metric"

// Under the DefaultAnalyzer the semconv rule is off, so no diagnostic fires
// even for a deviating name that would otherwise be flagged.
func register(m metric.Meter) {
	_, _ = m.Int64Counter("http.server.request.durations")
}
