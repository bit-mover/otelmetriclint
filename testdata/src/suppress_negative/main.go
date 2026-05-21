package suppress_negative

import "go.opentelemetry.io/otel/metric"

func f(m metric.Meter) {
	// nolint:otelmetriclint
	_, _ = m.Int64Counter("space_after_total") // want `total_suffix: counter name "space_after_total" must not end in _total` `unit_suffix: metric name "space_after_total" has unit-like suffix "total"`

	//NOLINT:otelmetriclint
	_, _ = m.Int64Counter("wrong_case_total") // want `total_suffix: counter name "wrong_case_total" must not end in _total` `unit_suffix: metric name "wrong_case_total" has unit-like suffix "total"`

	//nolint:gocritic
	_, _ = m.Int64Counter("wrong_linter_total") // want `total_suffix: counter name "wrong_linter_total" must not end in _total` `unit_suffix: metric name "wrong_linter_total" has unit-like suffix "total"`

	//nolint:otelmetriclintfoo
	_, _ = m.Int64Counter("substring_match_total") // want `total_suffix: counter name "substring_match_total" must not end in _total` `unit_suffix: metric name "substring_match_total" has unit-like suffix "total"`
}
