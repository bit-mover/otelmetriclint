package main

import (
	"reflect"
	"testing"
)

func TestParseOurFlags(t *testing.T) {
	cases := []struct {
		name     string
		argv     []string
		wantPath string
		wantVer  bool
		wantArgs []string
	}{
		{
			name:     "no flags",
			argv:     []string{"otelmetriclint", "./..."},
			wantArgs: []string{"otelmetriclint", "./..."},
		},
		{
			name:     "single-dash config with equals",
			argv:     []string{"otelmetriclint", "-config=foo.yaml", "./..."},
			wantPath: "foo.yaml",
			wantArgs: []string{"otelmetriclint", "./..."},
		},
		{
			name:     "single-dash config with space",
			argv:     []string{"otelmetriclint", "-config", "foo.yaml", "./..."},
			wantPath: "foo.yaml",
			wantArgs: []string{"otelmetriclint", "./..."},
		},
		{
			name:     "double-dash config with equals",
			argv:     []string{"otelmetriclint", "--config=foo.yaml", "./..."},
			wantPath: "foo.yaml",
			wantArgs: []string{"otelmetriclint", "./..."},
		},
		{
			name:     "double-dash config with space",
			argv:     []string{"otelmetriclint", "--config", "foo.yaml", "./..."},
			wantPath: "foo.yaml",
			wantArgs: []string{"otelmetriclint", "./..."},
		},
		{
			name:     "single-dash version",
			argv:     []string{"otelmetriclint", "-version"},
			wantVer:  true,
			wantArgs: []string{"otelmetriclint"},
		},
		{
			name:     "double-dash version",
			argv:     []string{"otelmetriclint", "--version"},
			wantVer:  true,
			wantArgs: []string{"otelmetriclint"},
		},
		{
			name:     "leaves -V and other flags for singlechecker",
			argv:     []string{"otelmetriclint", "-V=full", "--config=foo.yaml", "-other", "./..."},
			wantPath: "foo.yaml",
			wantArgs: []string{"otelmetriclint", "-V=full", "-other", "./..."},
		},
		{
			name:     "double-dash separator stops parsing",
			argv:     []string{"otelmetriclint", "--", "--config", "x", "--version"},
			wantArgs: []string{"otelmetriclint", "--", "--config", "x", "--version"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			argv := append([]string(nil), tc.argv...)
			gotPath, gotVer := parseOurFlags(&argv)
			if gotPath != tc.wantPath {
				t.Errorf("path: got %q want %q", gotPath, tc.wantPath)
			}
			if gotVer != tc.wantVer {
				t.Errorf("version: got %v want %v", gotVer, tc.wantVer)
			}
			if !reflect.DeepEqual(argv, tc.wantArgs) {
				t.Errorf("remaining: got %v want %v", argv, tc.wantArgs)
			}
		})
	}
}
