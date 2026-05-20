# otelmetriclint

A Go static analyzer for OpenTelemetry metric names. Catches the OTel→Prometheus export trap: forbids `_total` suffix on counters, forbids unit suffixes in names, requires `metric.WithUnit(...)` on histograms. Also enforces snake_case structure and an optional prefix allowlist.

Built on `golang.org/x/tools/go/analysis`. Runs standalone via `otelmetriclint ./...` or via `go vet -vettool=$(which otelmetriclint) ./...`.

Status: `v0.3.0` is the current release. The analyzer is usable today, but still pre-1.0; rule defaults and recognizer behavior can tighten as more OpenTelemetry metric patterns are covered.

## Install

```bash
go install github.com/bit-mover/otelmetriclint/cmd/otelmetriclint@latest
```

For Go 1.24+ projects that want to track the linter as a tool dependency in `go.mod`:

```bash
go get -tool github.com/bit-mover/otelmetriclint/cmd/otelmetriclint@v0.3.0
go tool otelmetriclint ./...
```

Or download a pre-built binary for your platform from the [releases page](https://github.com/bit-mover/otelmetriclint/releases).

## Usage

```bash
otelmetriclint ./...
otelmetriclint -config=path/to/.otelmetriclint.yaml ./...
```

Without `-config`, the tool looks for `.otelmetriclint.yaml` in the current working directory; if absent, it falls back to built-in defaults.

## Rules

| ID | Default | Catches |
| --- | --- | --- |
| `string_literal` | on | Name argument is not a string literal — can't statically check anything else |
| `structural` | on | Not snake_case / contains uppercase / leading digit / `..` / `__` |
| `prefix` | **off** | First segment not in `prefix.allowed` (rule is no-op when allowed list is empty) |
| `total_suffix` | on | Counter ends in `_total` (Prom exporter appends `_total`, causing double-suffix) |
| `unit_suffix` | on | Final segment is a UCUM unit code like `seconds`, `bytes`, `ms` — units belong in `WithUnit(...)`. Quantity descriptors (`duration`, `count`) are allowed because OTel semconv uses them canonically (`http.server.request.duration`, `db.client.connection.count`). |
| `histogram_unit` | on | Histogram created without `metric.WithUnit(...)` |

## Configuration

See [.otelmetriclint.yaml](./.otelmetriclint.yaml) for an annotated example. All fields are optional; user config merges over defaults. Unknown fields error at load to catch typos.

## Known limitations

- **Closure-call-site analysis is limited.** Options passed inside a closure body are invisible at a later call through a function-typed local variable. For example, if `histogram := func(...) metric.Float64Histogram { ... metric.WithUnit("s") ... }` is later called as `histogram("foo", "bar")`, the call site is recognized as the metric-creation site but the defaulted `WithUnit("s")` option inside the closure body is not visible there. This can produce false-positive `histogram_unit` diagnostics. Workaround: inline the metric construction at the call site so options are statically visible, or use an inline suppression once `otelmetriclint:disable:histogram_unit` support lands.

## Detected call sites

The recognizer is type-driven: any call whose result type matches a `go.opentelemetry.io/otel/metric.*` instrument interface is treated as a metric-creation call. This catches:

- SDK methods (`meter.Int64Counter("name")`, etc.) — return tuple `(Int64Counter, error)`
- Must-style wrapper helpers (`MustInt64Counter(meter, "name")`) — return the instrument directly
- Generic helpers — `pass.TypesInfo` resolves return types correctly

No per-project configuration is needed for standard wrapper patterns. Use `helpers:` config only for wrappers where the metric name is not the first string argument.

## Why these specific rules

See the [OTel→Prometheus compatibility spec](https://opentelemetry.io/docs/specs/otel/compatibility/prometheus_and_openmetrics/) and the [OTel metric semantic conventions](https://opentelemetry.io/docs/specs/semconv/general/metrics/). Short version:

- The Prometheus exporter appends `_total` to counters and the unit from `WithUnit(...)` to all instruments. Source-side names that already contain `_total` or a UCUM unit code produce double-encoded names like `foo_total_total` or `foo_seconds_seconds`. (Note: `foo_duration_seconds` is *not* a double-encode — `duration` is a quantity descriptor, not a unit; the Prom-side `_duration_seconds` is the canonical convention.)
- Histograms in particular need explicit `WithUnit(...)` once unit suffixes are forbidden in the name.

## Credit

Visitor structure inspired by [`promlinter`](https://github.com/yeya24/promlinter) (Apache-2.0). Rule bodies are written fresh against OTel semantic conventions; no code is copied.

## License

Apache-2.0 — see [LICENSE](./LICENSE).
