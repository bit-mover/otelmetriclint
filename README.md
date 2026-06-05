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
| `cross_package_uniqueness` | **off** | The same OTel metric name is registered in more than one package (detected across import edges — see limitations below) |
| `pluralization` | on | UpDownCounter (and ObservableUpDownCounter) name's leaf token looks pluralized (e.g. `...connections`) — UpDownCounters measure a current value, so prefer a singular noun like `connection.count`. Heuristic: leaf ends in `s` (excluding `ss`/`us`/`is`), length >= 4, minus an allowlist. |
| `ucum_unit` | on | `metric.WithUnit(...)` value is not a valid UCUM case-sensitive code (e.g. `"seconds"` instead of `"s"`, `"bytes"` instead of `"By"`). Suggests the canonical UCUM code when a known verbose English word is detected. Extend the accept list with `ucum_unit.additional_allow` for project-local conventions. |
| `semconv` | **off** | Metric name's root segment is a claimed OTel semconv namespace (e.g. `http`, `db`, `k8s`) but the full name is not a registered semantic-convention metric. Suggests the closest registered name when the edit distance is small (≤ 3 and < half the name length). Unclaimed root segments (project-specific prefixes like `cedar`, `acme`) are never flagged — the namespace firewall prevents false positives. |

### cross_package_uniqueness

Enable this rule when you want the linter to flag the same OTel metric name being registered in more than one package. Because duplicate names across independently-deployed binaries cannot be detected statically, the rule is **off by default**.

To enable:

```yaml
rules:
  cross_package_uniqueness: true
```

**Limitations:**

1. **Import-edge reach.** Collisions are only detected when the registrations are reachable through an import edge — e.g. a package that imports another, or a `cmd` package that imports both. Two fully independent binaries that happen to define the same metric name are not compared.

2. **Diagnostic anchor.** The diagnostic is reported at the importing package's registration site, not symmetrically at both sites. The other side of the collision is identified in the message, but is not itself flagged.

### pluralization

This rule is **on by default**. It flags UpDownCounter and ObservableUpDownCounter registrations whose leaf name segment looks pluralized — e.g. `system.network.connections` — because UpDownCounters measure a current value (a gauge-like quantity) and should use a singular noun.

The heuristic is: the leaf segment ends in `s`, is at least 4 characters long, and does not end in `ss`, `us`, or `is`. A built-in allowlist exempts words that are legitimately singular despite ending in `s`:

- `series`
- `kubernetes`
- `https`
- `analytics`
- `statistics`
- `diagnostics`

To extend the allowlist with project-specific terms, add them under `pluralization.additional_allow` in your config file. This **appends** to the built-in list — it does not replace it:

```yaml
pluralization:
  additional_allow: [myseries, customterm]
```

### semconv

This rule is **off by default**. When enabled, it flags any metric name whose root namespace segment is a claimed OTel semconv namespace (e.g. `http`, `db`, `k8s`, `process`, `system`) but whose full name is not a registered semantic-convention metric. This catches typos and near-misses like `http.server.request.durations` instead of the canonical `http.server.request.duration`.

To enable:

```yaml
rules:
  semconv: true
```

**Namespace firewall.** Only metric names whose root segment is a known semconv namespace are checked. Project-specific prefixes (e.g. `cedar.policy.swaps`, `acme.billing.invoices`) are never flagged, even when the rule is enabled. This prevents the rule from producing false positives on custom metric namespaces.

**Near-miss suggestions.** When a deviating name is within Levenshtein distance ≤ 3 of a registered name in the same namespace, and that distance is strictly less than half the name's length, the diagnostic includes a `did you mean "..."?` suggestion.

**Allow list.** To accept non-registered names under a claimed namespace (e.g. internal metrics that share an OTel namespace prefix), add them under `semconv.additional_allow` in your config file. This **appends** to the default empty list:

```yaml
semconv:
  additional_allow: [http.server.custom.internal_metric]
```

## Configuration

See [.otelmetriclint.yaml](./.otelmetriclint.yaml) for an annotated example. All fields are optional; user config merges over defaults. Unknown fields error at load to catch typos.

## Suppressing diagnostics

Standalone runs honor a subset of golangci-lint's `//nolint` directive grammar. Whole-linter suppression only — no per-rule scoping. The same comments that work under golangci-lint also work when running this analyzer directly.

```go
// trailing on the call's start line
_, _ = m.Int64Counter("legacy_total") //nolint:otelmetriclint

// immediately above the call (adjacent — no blank line)
//nolint:otelmetriclint
_, _ = m.Int64Counter("legacy_total")

// above an enclosing FuncDecl or FuncLit (covers the whole body)
//nolint:otelmetriclint
func registerLegacyMetrics(m metric.Meter) {
    _, _ = m.Int64Counter("a_total")
    _, _ = m.Int64Counter("b_total")
}

// above `package` (covers the whole file)
//nolint:otelmetriclint
package legacy
```

When run under golangci-lint, the runner applies its own nolint filtering — both filter the same diagnostics, so the double-filter is a no-op.

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
