package good

import "go.opentelemetry.io/otel/metric"

// ucum_units exercises metric creation call sites that carry valid UCUM unit
// codes. The recognizer debug analyzer (recognizer_test.go) sees these and
// emits a "found <Kind> name=..." diagnostic — each annotation below asserts
// that the recognizer picks up the call site correctly.
//
// These calls do NOT test ucum_unit rule silence; that is covered by the
// testdata/src/bad_ucum_unit/ fixtures (valid unit lines there carry no annotation).

func validUCUMUnits(m metric.Meter) {
	_, _ = m.Float64Histogram("http.server.request.body.size", metric.WithUnit("By"))  // want `found Float64Histogram name="http.server.request.body.size"`
	_, _ = m.Float64Histogram("process.cpu.time", metric.WithUnit("s"))                // want `found Float64Histogram name="process.cpu.time"`
	_, _ = m.Int64Counter("network.io.op", metric.WithUnit("{packet}"))                // want `found Int64Counter name="network.io.op"`
	_, _ = m.Float64Gauge("disk.io.rate", metric.WithUnit("By/s"))                    // want `found Float64Gauge name="disk.io.rate"`
}
