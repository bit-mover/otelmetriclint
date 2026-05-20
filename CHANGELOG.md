# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
