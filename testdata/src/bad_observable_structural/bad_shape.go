package bad_observable_structural

import "go.opentelemetry.io/otel/metric"

// Observable instruments with structurally invalid names fire the structural
// rule just like synchronous instruments do.
func observableStructuralBad(m metric.Meter) {
	_, _ = m.Int64ObservableCounter("Bad.name")           // want `structural: metric name "Bad.name" must be ASCII-lowercase`
	_, _ = m.Float64ObservableGauge("obs..double_dot")    // want `structural: metric name "obs..double_dot" contains consecutive dot delimiters`
}
