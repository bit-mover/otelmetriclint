# otelmetriclint

A Go static analyzer for OpenTelemetry metric names. Catches the OTel→Prometheus export trap: forbids `_total` suffix on counters, forbids unit suffixes in names, requires `metric.WithUnit(...)` on histograms. Also enforces snake_case structure and an optional prefix allowlist.

Built on `golang.org/x/tools/go/analysis`. Runs standalone via `otelmetriclint ./...` or via `go vet -vettool=$(which otelmetriclint) ./...`.

## Why this exists

I migrated a service from Prometheus to OpenTelemetry metrics and found it surprisingly easy to get wrong: names that worked fine under the Prometheus client got double-encoded by the OTel→Prom exporter (`_total_total`, `_seconds_seconds`), units belonged in `WithUnit(...)` instead of the name, and the failure mode was silent — metrics shipped, dashboards just broke. I looked for a linter that would catch these at build time and didn't find one, so I wrote this.

## Install

```bash
go install github.com/bit-mover/otelmetriclint/cmd/otelmetriclint@latest
```

Or, with Go 1.24+, track it as a tool dependency in your project's `go.mod` so Renovate / Dependabot can auto-update it:

```bash
go get -tool github.com/bit-mover/otelmetriclint/cmd/otelmetriclint@latest
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

- **Closure-call-site analysis is limited.** Wrapper-body suppression (v0.3.0) only catches calls inside a function or closure whose return type is a metric instrument. If a wrapper is invoked through a function-typed local variable — e.g. `histogram := func(name string, opts ...metric.Float64HistogramOption) metric.Float64Histogram { ... meter.Float64Histogram(name, append(opts, metric.WithUnit("s"))...) }` and then `histogram("foo")` — the recognizer still sees the call site as a metric-creation site, and options set inside the closure body (like `WithUnit("s")`) are invisible. This typically produces false-positive `histogram_unit` diagnostics. Workaround: inline the construction at the call site so the option is statically visible.

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
