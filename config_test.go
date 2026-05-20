package otelmetriclint

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	c := Default()
	if !c.Rules["structural"] {
		t.Error("structural should default true")
	}
	if c.Rules["prefix"] {
		t.Error("prefix should default false")
	}
	if len(c.UnitSuffix.Forbidden) == 0 {
		t.Error("default unit_suffix.forbidden should be non-empty")
	}
}

func TestLoadConfigMergesOverDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".otelmetriclint.yaml")
	contents := []byte(`
rules:
  prefix: true
prefix:
  allowed: [iam, customers]
`)
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !c.Rules["prefix"] {
		t.Error("user override of rules.prefix not applied")
	}
	if !c.Rules["structural"] {
		t.Error("default rules.structural lost after merge")
	}
	wantAllowed := []string{"iam", "customers"}
	if !reflect.DeepEqual(c.Prefix.Allowed, wantAllowed) {
		t.Errorf("Prefix.Allowed = %v, want %v", c.Prefix.Allowed, wantAllowed)
	}
}

func TestLoadConfigRejectsUnknownFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".otelmetriclint.yaml")
	if err := os.WriteFile(path, []byte("typo_field: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("expected error for unknown field, got nil")
	}
}

func TestLoadConfigEmptyPathReturnsDefaults(t *testing.T) {
	c, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c, Default()) {
		t.Error("Load(\"\") should equal Default()")
	}
}

func TestLoadConfigEmptyFileReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".otelmetriclint.yaml")
	// All-comment file — yaml.Decode returns io.EOF, which should be treated as
	// "no overrides" so the shipped example config works out of the box.
	if err := os.WriteFile(path, []byte("# only comments\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c, Default()) {
		t.Error("Load of comment-only file should equal Default()")
	}
}
