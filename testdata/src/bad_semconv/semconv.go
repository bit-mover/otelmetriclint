package bad_semconv

import "go.opentelemetry.io/otel/metric"

func register(m metric.Meter) {
	// Deviating name: near-miss on the canonical "http.server.request.duration".
	_, _ = m.Int64Counter("http.server.request.durations") // want `semconv: metric name "http\.server\.request\.durations" is not a registered OpenTelemetry semantic-convention metric in the "http" namespace; did you mean "http\.server\.request\.duration"\?`

	// Registered canonical name: must stay silent even when the rule is enabled.
	_, _ = m.Int64Counter("db.client.connection.count")

	// Project-namespaced name: "cedar" is not a claimed semconv namespace,
	// so the rule must stay silent regardless of whether the name is registered.
	_, _ = m.Int64Counter("cedar.policy.swaps")
}
