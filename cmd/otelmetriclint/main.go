// Command otelmetriclint validates OpenTelemetry metric instrument
// creation call sites against a configurable rule set.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/bit-mover/otelmetriclint"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	// Register on flag.CommandLine for -h discoverability only; the real
	// parsing happens in parseOurFlags before singlechecker.Main runs.
	flag.String("config", "", "path to YAML config (defaults: ./.otelmetriclint.yaml then built-in defaults)")
	flag.Bool("version", false, "print version and exit")

	argv := os.Args
	path, showVersion := parseOurFlags(&argv)
	os.Args = argv

	if showVersion {
		fmt.Println("otelmetriclint", version)
		return
	}
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

// parseOurFlags extracts -config and -version (each accepting one or two
// dashes, with `=` or space-separated values) from argv, leaving the rest
// for singlechecker.Main to parse. The `--` separator stops parsing.
func parseOurFlags(argv *[]string) (configPath string, showVersion bool) {
	in := *argv
	if len(in) == 0 {
		return "", false
	}
	out := make([]string, 0, len(in))
	out = append(out, in[0])
	for i := 1; i < len(in); i++ {
		a := in[i]
		if a == "--" {
			out = append(out, in[i:]...)
			break
		}
		name, value, hasValue := splitFlag(a)
		switch name {
		case "config":
			if hasValue {
				configPath = value
			} else if i+1 < len(in) {
				i++
				configPath = in[i]
			}
		case "version":
			showVersion = true
		default:
			out = append(out, a)
		}
	}
	*argv = out
	return configPath, showVersion
}

func splitFlag(arg string) (name, value string, hasValue bool) {
	if len(arg) < 2 || arg[0] != '-' {
		return "", "", false
	}
	s := strings.TrimPrefix(strings.TrimPrefix(arg, "-"), "-")
	if s == "" {
		return "", "", false
	}
	if n, v, ok := strings.Cut(s, "="); ok {
		return n, v, true
	}
	return s, "", false
}
