package bad_prefix

import "go.opentelemetry.io/otel/metric"

func rogue(m metric.Meter) {
	_, _ = m.Int64Counter("rogue.thing") // want `prefix: metric name "rogue.thing" has prefix "rogue" not in the allowlist`
}
