// Package otelmetriclint contains an analysis.Analyzer that validates
// OpenTelemetry metric instrument creation call sites against a
// configurable rule set covering snake_case structure, allowed prefixes,
// the OTel→Prometheus export trap (no _total, no unit suffix in name),
// and the requirement that histograms set metric.WithUnit(...).
//
// See https://github.com/bit-mover/otelmetriclint for usage.
package otelmetriclint
