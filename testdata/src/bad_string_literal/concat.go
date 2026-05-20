package bad_string_literal

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

func dynamic(m metric.Meter, suffix string) {
	_, _ = m.Int64Counter(fmt.Sprintf("foo.%s", suffix)) // want `found Int64Counter name=""`
}
