package bad_string_literal

import "go.opentelemetry.io/otel/metric"

func fromVar(m metric.Meter, name string) {
	_, _ = m.Int64Counter(name) // want `found Int64Counter name=""`
}
