package otelmetriclint

import (
	"go/types"

	"github.com/bit-mover/otelmetriclint/rules"
)

const otelMetricPkgPath = "go.opentelemetry.io/otel/metric"

// matchInstrumentKind reports the rules.InstrumentKind corresponding to t
// if t is one of the metric instrument interfaces from
// go.opentelemetry.io/otel/metric. Returns (KindUnknown, false) otherwise.
func matchInstrumentKind(t types.Type) (rules.InstrumentKind, bool) {
	named, ok := t.(*types.Named)
	if !ok {
		return rules.KindUnknown, false
	}
	obj := named.Obj()
	if obj.Pkg() == nil || obj.Pkg().Path() != otelMetricPkgPath {
		return rules.KindUnknown, false
	}
	switch obj.Name() {
	case "Int64Counter":
		return rules.KindInt64Counter, true
	case "Float64Counter":
		return rules.KindFloat64Counter, true
	case "Int64UpDownCounter":
		return rules.KindInt64UpDownCounter, true
	case "Float64UpDownCounter":
		return rules.KindFloat64UpDownCounter, true
	case "Int64Histogram":
		return rules.KindInt64Histogram, true
	case "Float64Histogram":
		return rules.KindFloat64Histogram, true
	case "Int64Gauge":
		return rules.KindInt64Gauge, true
	case "Float64Gauge":
		return rules.KindFloat64Gauge, true
	}
	return rules.KindUnknown, false
}
