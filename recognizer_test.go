package otelmetriclint

import (
	"testing"

	_ "go.opentelemetry.io/otel/metric" // ensure module graph keeps otel/metric so packages.Load can resolve it
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/packages"

	"github.com/bit-mover/otelmetriclint/rules"
)

func TestMatchInstrumentKindFromMetricPackage(t *testing.T) {
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedTypesInfo}
	pkgs, err := packages.Load(cfg, "go.opentelemetry.io/otel/metric")
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if len(pkgs[0].Errors) > 0 {
		t.Fatalf("package load errors: %v", pkgs[0].Errors)
	}
	scope := pkgs[0].Types.Scope()
	cases := []struct {
		typeName string
		want     rules.InstrumentKind
	}{
		{"Int64Counter", rules.KindInt64Counter},
		{"Float64Counter", rules.KindFloat64Counter},
		{"Int64UpDownCounter", rules.KindInt64UpDownCounter},
		{"Float64UpDownCounter", rules.KindFloat64UpDownCounter},
		{"Int64Histogram", rules.KindInt64Histogram},
		{"Float64Histogram", rules.KindFloat64Histogram},
		{"Int64Gauge", rules.KindInt64Gauge},
		{"Float64Gauge", rules.KindFloat64Gauge},
	}
	for _, tc := range cases {
		t.Run(tc.typeName, func(t *testing.T) {
			obj := scope.Lookup(tc.typeName)
			if obj == nil {
				t.Fatalf("%s not found in otel/metric scope", tc.typeName)
			}
			got, ok := matchInstrumentKind(obj.Type())
			if !ok {
				t.Fatalf("matchInstrumentKind returned ok=false")
			}
			if got != tc.want {
				t.Errorf("matchInstrumentKind(%s) = %v, want %v", tc.typeName, got, tc.want)
			}
		})
	}
}

func TestMatchInstrumentKindNonMetricType(t *testing.T) {
	// A type from outside otel/metric should return KindUnknown, false.
	cfg := &packages.Config{Mode: packages.NeedTypes}
	pkgs, err := packages.Load(cfg, "errors")
	if err != nil {
		t.Fatal(err)
	}
	scope := pkgs[0].Types.Scope()
	obj := scope.Lookup("New") // a *types.Func, not a Named type — exercise the non-Named branch
	if obj == nil {
		t.Fatal("errors.New not found")
	}
	kind, ok := matchInstrumentKind(obj.Type())
	if ok {
		t.Errorf("expected ok=false for errors.New, got kind=%v", kind)
	}
}

// debugAnalyzer records every MetricCall the recognizer finds. Used to
// unit-test findMetricCalls in isolation, separate from the rule pipeline.
func debugAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "recognizerdebug",
		Doc:  "records every MetricCall",
		Run: func(pass *analysis.Pass) (interface{}, error) {
			for _, c := range findMetricCalls(pass, nil) {
				pass.Reportf(c.Pos, "found %s name=%q", c.Kind, c.Name)
			}
			return nil, nil
		},
	}
}

func TestFindMetricCallsRecognizesSDKAndWrappers(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), debugAnalyzer(), "./src/good")
}
