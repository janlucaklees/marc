package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir_usesXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdgtest")
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}
	want := filepath.Join("/tmp/xdgtest", "marc")
	if dir != want {
		t.Errorf("ConfigDir() = %q, want %q", dir, want)
	}
}

func TestConfigDir_fallsBackToHomeConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/tmp/hometest")
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}
	want := filepath.Join("/tmp/hometest", ".config", "marc")
	if dir != want {
		t.Errorf("ConfigDir() = %q, want %q", dir, want)
	}
}

func TestLoadConfig_missingFileReturnsZeroValue(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg != (Config{}) {
		t.Errorf("LoadConfig() = %+v, want zero value", cfg)
	}
}

func TestLoadConfig_parsesBothFields(t *testing.T) {
	dir := t.TempDir()
	content := "pandoc_bin = \"/usr/bin/pandoc\"\nchromium_bin = \"/usr/bin/chromium\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing test config: %v", err)
	}
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	want := Config{PandocBin: "/usr/bin/pandoc", ChromiumBin: "/usr/bin/chromium"}
	if cfg != want {
		t.Errorf("LoadConfig() = %+v, want %+v", cfg, want)
	}
}

func TestLoadConfig_partialFieldsLeaveOthersZero(t *testing.T) {
	dir := t.TempDir()
	content := "pandoc_bin = \"/usr/bin/pandoc\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing test config: %v", err)
	}
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	want := Config{PandocBin: "/usr/bin/pandoc"}
	if cfg != want {
		t.Errorf("LoadConfig() = %+v, want %+v", cfg, want)
	}
}

func TestLoadConfig_malformedTOMLReturnsError(t *testing.T) {
	dir := t.TempDir()
	content := "this is not valid toml = = =\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing test config: %v", err)
	}
	if _, err := LoadConfig(dir); err == nil {
		t.Error("LoadConfig() error = nil, want error for malformed TOML")
	}
}
