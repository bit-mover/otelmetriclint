package bad_structural

import "go.opentelemetry.io/otel/metric"

func doubled(m metric.Meter) {
	_, _ = m.Int64Counter("foo..bar") // want `structural: metric name "foo..bar" contains consecutive dot delimiters`
	_, _ = m.Int64Counter("foo__bar") // want `structural: metric name "foo__bar" contains consecutive underscore delimiters`
}
