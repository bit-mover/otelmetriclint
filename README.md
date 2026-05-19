# otelmetriclint

A Go static analyzer for OpenTelemetry metric names. Catches the OTel→Prometheus export trap: forbids `_total` suffix on counters, forbids unit suffixes in names, requires `metric.WithUnit(...)` on histograms. Also enforces snake_case structure and an optional prefix allowlist.

Built on `golang.org/x/tools/go/analysis`. Runs standalone via `otelmetriclint ./...` or via `go vet -vettool=$(which otelmetriclint) ./...`.

## Status

Pre-release. Tracking toward `v0.1.0`.

## Install

```bash
go install github.com/bit-mover/otelmetriclint/cmd/otelmetriclint@latest
```

Or download a pre-built binary from the [releases page](https://github.com/bit-mover/otelmetriclint/releases).

## Credit

Visitor structure inspired by [`promlinter`](https://github.com/yeya24/promlinter) (Apache-2.0). Rule bodies are written fresh against OTel semantic conventions; no code is copied.

## License

Apache-2.0 — see [LICENSE](./LICENSE).
