package otelmetriclint

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Config controls which rules run and how. Unknown YAML fields are
// rejected at load to catch typos.
type Config struct {
	Rules         map[string]bool     `yaml:"rules,omitempty"`
	Prefix        PrefixConfig        `yaml:"prefix,omitempty"`
	UnitSuffix    UnitSuffixConfig    `yaml:"unit_suffix,omitempty"`
	Pluralization PluralizationConfig `yaml:"pluralization,omitempty"`
	UCUMUnit      UCUMUnitConfig      `yaml:"ucum_unit,omitempty"`
	Semconv       SemconvConfig       `yaml:"semconv,omitempty"`
	Helpers       []HelperConfig      `yaml:"helpers,omitempty"`
}

// PrefixConfig holds the allowlist for the prefix rule.
type PrefixConfig struct {
	Allowed []string `yaml:"allowed,omitempty"`
}

// UnitSuffixConfig holds the forbidden-suffix list for the unit_suffix rule.
type UnitSuffixConfig struct {
	Forbidden  []string `yaml:"forbidden,omitempty"`
	Additional []string `yaml:"additional,omitempty"`
}

// PluralizationConfig holds the allowlist for the pluralization rule.
type PluralizationConfig struct {
	AdditionalAllow []string `yaml:"additional_allow,omitempty"`
}

// UCUMUnitConfig holds the allowlist for the ucum_unit rule.
type UCUMUnitConfig struct {
	AdditionalAllow []string `yaml:"additional_allow,omitempty"`
}

// SemconvConfig holds the allowlist for the semconv rule.
type SemconvConfig struct {
	AdditionalAllow []string `yaml:"additional_allow,omitempty"`
}

// HelperConfig declares an override for a wrapper where the metric name
// is not the first string argument.
type HelperConfig struct {
	Pkg     string `yaml:"pkg"`
	Func    string `yaml:"func"`
	NameArg int    `yaml:"name_arg"`
}

// Default returns the built-in default Config. All v1 rules are on except
// prefix (empty allowlist = no constraint).
func Default() Config {
	return Config{
		Rules: map[string]bool{
			"string_literal":            true,
			"structural":                true,
			"prefix":                    false,
			"total_suffix":              true,
			"unit_suffix":               true,
			"histogram_unit":            true,
			"cross_package_uniqueness":  false,
			"pluralization":             true,
			"ucum_unit":                 true,
			"semconv":                   false,
		},
		UnitSuffix: UnitSuffixConfig{
			// UCUM unit codes and their expansions only. Quantity
			// descriptors like `duration`, `count`, `size`,
			// `utilization` are NOT included — OTel semconv uses them
			// canonically (`http.server.request.duration`,
			// `db.client.connection.count`). Including them would
			// flag canonical OTel names.
			Forbidden: []string{
				"seconds", "bytes", "ms", "us", "ns", "s",
				"kb", "mb", "gb", "b", "total",
			},
		},
	}
}

// Load reads YAML config from path (or returns Default() if path is "").
// Unknown fields are rejected.
func Load(path string) (Config, error) {
	c := Default()
	if path == "" {
		return c, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer func() { _ = f.Close() }()
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	var user Config
	if err := dec.Decode(&user); err != nil {
		// An empty or all-comment file is logically "no overrides" — keep defaults.
		if errors.Is(err, io.EOF) {
			return c, nil
		}
		return Config{}, fmt.Errorf("decode config %s: %w", path, err)
	}
	merge(&c, user)
	return c, nil
}

// merge overlays user's settings on top of base. The Rules map is unioned;
// top-level fields are replaced wholesale.
func merge(base *Config, user Config) {
	for k, v := range user.Rules {
		if base.Rules == nil {
			base.Rules = make(map[string]bool)
		}
		base.Rules[k] = v
	}
	if len(user.Prefix.Allowed) > 0 {
		base.Prefix.Allowed = user.Prefix.Allowed
	}
	if len(user.UnitSuffix.Forbidden) > 0 {
		base.UnitSuffix.Forbidden = user.UnitSuffix.Forbidden
	}
	if len(user.UnitSuffix.Additional) > 0 {
		base.UnitSuffix.Additional = append(base.UnitSuffix.Additional, user.UnitSuffix.Additional...)
	}
	if len(user.Pluralization.AdditionalAllow) > 0 {
		base.Pluralization.AdditionalAllow = append(base.Pluralization.AdditionalAllow, user.Pluralization.AdditionalAllow...)
	}
	if len(user.UCUMUnit.AdditionalAllow) > 0 {
		base.UCUMUnit.AdditionalAllow = append(base.UCUMUnit.AdditionalAllow, user.UCUMUnit.AdditionalAllow...)
	}
	if len(user.Semconv.AdditionalAllow) > 0 {
		base.Semconv.AdditionalAllow = append(base.Semconv.AdditionalAllow, user.Semconv.AdditionalAllow...)
	}
	if len(user.Helpers) > 0 {
		base.Helpers = user.Helpers
	}
}
