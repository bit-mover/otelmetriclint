// Command otelmetriclint validates OpenTelemetry metric instrument
// creation call sites against a configurable rule set.
package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/bit-mover/otelmetriclint"
)

func main() {
	configPath := flag.String("config", "", "path to YAML config (defaults: ./.otelmetriclint.yaml then built-in defaults)")
	flag.Parse()

	path := *configPath
	if path == "" {
		if _, err := os.Stat(".otelmetriclint.yaml"); err == nil {
			path = ".otelmetriclint.yaml"
		}
	}
	cfg, err := otelmetriclint.Load(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config:", err)
		os.Exit(2)
	}
	singlechecker.Main(otelmetriclint.New(cfg))
}
