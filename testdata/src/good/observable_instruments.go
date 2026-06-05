package good

import "go.opentelemetry.io/otel/metric"

// Observable (asynchronous) instruments — all six variants should be
// recognized by the recognizer without any diagnostics firing.

func observableInstruments(m metric.Meter) {
	_, _ = m.Int64ObservableCounter("good.obs_counter")                                              // want `found Int64ObservableCounter name="good.obs_counter"`
	_, _ = m.Float64ObservableCounter("good.float_obs_counter")                                      // want `found Float64ObservableCounter name="good.float_obs_counter"`
	_, _ = m.Int64ObservableUpDownCounter("good.obs_updown")                                         // want `found Int64ObservableUpDownCounter name="good.obs_updown"`
	_, _ = m.Float64ObservableUpDownCounter("good.float_obs_updown")                                 // want `found Float64ObservableUpDownCounter name="good.float_obs_updown"`
	_, _ = m.Int64ObservableGauge("good.obs_gauge")                                                  // want `found Int64ObservableGauge name="good.obs_gauge"`
	_, _ = m.Float64ObservableGauge("good.float_obs_gauge")                                          // want `found Float64ObservableGauge name="good.float_obs_gauge"`
}
