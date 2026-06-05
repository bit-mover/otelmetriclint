# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- **Const-folded metric names now resolve correctly (closes #16).** Previously, a metric name built from a chain of named constants (e.g. `const prefix = "app."; const name = prefix + "requests"`) was not evaluated to its final string value, causing the recognizer to treat it as a non-literal and emit a false-positive `string_literal` diagnostic. The analyzer now folds constant expressions to their concrete string value before applying rules, so names constructed entirely from `const` declarations are accepted without suppression.

- **Observable instruments are now recognized (closes #15).** `Int64ObservableCounter`, `Float64ObservableCounter`, `Int64ObservableUpDownCounter`, `Float64ObservableUpDownCounter`, `Int64ObservableGauge`, and `Float64ObservableGauge` meter methods were missing from the recognizer's instrument table and were silently skipped. All six are now treated as metric-creation call sites. `ObservableCounter` variants are classified as monotonic (equivalent to `Counter`) so that the `total_suffix` rule correctly requires a `_total` suffix, consistent with Prometheus naming conventions for monotonically increasing counters.

### Added

- **`cross_package_uniqueness` rule (off by default).** Flags the same OTel metric name registered in more than one package, detected across import edges. Off by default because detection is limited to packages reachable through a shared import graph; two fully independent binaries are not compared. Enable with `rules: cross_package_uniqueness: true`. Closes #17.
- **Inline suppression directives.** Standalone runs now honor `//nolint:otelmetriclint` placed trailing on a call's line, on the line immediately above a call, above an enclosing `func` or closure, or above the `package` keyword. Mirrors the subset of golangci-lint's nolint grammar that whole-linter suppression covers — no per-rule scoping. See README §"Suppressing diagnostics". Closes #18.
- **`pluralization` rule (on by default).** Flags UpDownCounter and ObservableUpDownCounter metric names whose leaf token appears pluralized (e.g. `db.client.connections`); UpDownCounters represent a current value, so a singular form like `connection.count` is preferred. Uses a permissive heuristic (leaf ends in `s`, excluding `ss`/`us`/`is`, length >= 4) plus a built-in allowlist extendable via `pluralization: additional_allow: [...]`. Closes #29.

## [0.3.0] - 2026-05-20

### Changed

- **Recognizer skips helper-wrapper bodies.** A CallExpr whose enclosing function or closure returns one of the `go.opentelemetry.io/otel/metric.*` instrument interfaces is no longer flagged as a metric-creation call site. Helper wrappers like `func MustInt64Counter(meter, name, opts) metric.Int64Counter { return meter.Int64Counter(name, opts...) }` previously generated false-positive `string_literal` diagnostics on the inner SDK call (because `name` is a parameter, not a literal). The real metric-creation site is at the wrapper's *callers*, where the name is a literal — that's where rules now apply.

### Migration

If a project's `func` or closure returns a metric instrument and contains a metric-instrument call inside, the linter now correctly attributes the check to the wrapper's callers. No config changes needed. Projects that *want* the v0.2.0 behavior of inspecting wrapper bodies have no migration path — the suppression is structural.

### Known limitations

- Wrapper-body suppression only fires for calls *inside* a function or closure whose return type is a metric instrument. Wrappers invoked through a function-typed local variable (e.g. `histogram := func(...) metric.Float64Histogram { ... }` then `histogram("foo")`) still appear as metric-creation call sites, so options set inside the closure body — including `WithUnit(...)` — are invisible. Expect false-positive `histogram_unit` diagnostics in that shape. Workaround: inline the construction at the call site. See README "Known limitations".

[0.3.0]: https://github.com/bit-mover/otelmetriclint/releases/tag/v0.3.0

## [0.2.0] - 2026-05-20

### Changed

- **`unit_suffix` default forbidden list**: removed `duration` and `count`. Both are *quantity descriptors* (not units) and are used canonically by OTel semantic conventions (`http.server.request.duration`, `db.client.connection.count`, `process.thread.count`). The previous list flagged canonical OTel names as violations. The unit goes in `WithUnit(...)` — for duration histograms that's `WithUnit("s")`, which the Prometheus exporter then translates to `_seconds`, producing the canonical Prom-side name `foo_duration_seconds` with no double-encoding.
- README updated to clarify that `foo_duration_seconds` is NOT a double-encoded name (the README's earlier example was incorrect).

### Migration

Projects using the default config that previously had a few false-positive diagnostics on `.duration` / `.count`-suffixed names will see those diagnostics disappear. No config change needed. Users who explicitly want the stricter v0.1.0 behavior can set `unit_suffix.forbidden` to the old list in their config.

[0.2.0]: https://github.com/bit-mover/otelmetriclint/releases/tag/v0.2.0

## [0.1.0] - 2026-05-20

### Added

- Type-driven recognizer for OTel metric instrument creation call sites (SDK methods, Must-style wrappers, generic helpers)
- Six rules: `string_literal`, `structural`, `prefix`, `total_suffix`, `unit_suffix`, `histogram_unit`
- YAML config with defaults, merge, unknown-field rejection
- CLI binary via `singlechecker`
- `analysistest`-based test coverage for every rule
- Multi-platform binary releases via goreleaser

[0.1.0]: https://github.com/bit-mover/otelmetriclint/releases/tag/v0.1.0
