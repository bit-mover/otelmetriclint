# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-05-20

### Added

- Type-driven recognizer for OTel metric instrument creation call sites (SDK methods, Must-style wrappers, generic helpers)
- Six rules: `string_literal`, `structural`, `prefix`, `total_suffix`, `unit_suffix`, `histogram_unit`
- YAML config with defaults, merge, unknown-field rejection
- CLI binary via `singlechecker`
- `analysistest`-based test coverage for every rule
- Multi-platform binary releases via goreleaser

[0.1.0]: https://github.com/bit-mover/otelmetriclint/releases/tag/v0.1.0
