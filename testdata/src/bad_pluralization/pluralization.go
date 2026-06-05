package bad_pluralization

import "go.opentelemetry.io/otel/metric"

func pluralization(m metric.Meter) {
	// FLAGGED: Int64UpDownCounter with plural leaf "connections"
	_, _ = m.Int64UpDownCounter("db.client.connections") // want `pluralization: UpDownCounter name "db.client.connections" appears pluralized`

	// FLAGGED: Float64UpDownCounter with plural leaf "entries"
	_, _ = m.Float64UpDownCounter("cache.entries") // want `pluralization: UpDownCounter name "cache.entries" appears pluralized`

	// FLAGGED: Int64ObservableUpDownCounter with plural leaf "messages"
	_, _ = m.Int64ObservableUpDownCounter("queue.messages") // want `pluralization: UpDownCounter name "queue.messages" appears pluralized`

	// CLEAN: singular leaf "count" — no diagnostic
	_, _ = m.Int64UpDownCounter("db.client.connection.count")

	// CLEAN: leaf "status" ends in -us — excluded by heuristic, no diagnostic
	_, _ = m.Int64UpDownCounter("job.status")

	// CLEAN: leaf "series" is in the built-in allowlist — no diagnostic
	_, _ = m.Int64UpDownCounter("metrics.series")
}
