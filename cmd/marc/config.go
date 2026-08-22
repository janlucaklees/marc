package main

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// defaultConfigTOML is the starter config.toml shipped with marc,
// written out to the config dir on first run when none exists yet.
// See BootstrapDefaultConfig.
//
//go:embed config.example.toml
var defaultConfigTOML []byte

// Config holds optional overrides read from config.toml.
type Config struct {
	PandocBin   string `toml:"pandoc_bin"`
	ChromiumBin string `toml:"chromium_bin"`
}

// ConfigDir returns marc's config directory: $XDG_CONFIG_HOME/marc if
// XDG_CONFIG_HOME is set, otherwise $HOME/.config/marc.
func ConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "marc"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "marc"), nil
}

// LoadConfig reads <dir>/config.toml. If the file does not exist, it
// returns a zero-value Config and no error. Any other read or parse
// error is returned as-is.
func LoadConfig(dir string) (Config, error) {
	var cfg Config
	path := filepath.Join(dir, "config.toml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// BootstrapDefaultConfig writes the embedded starter config.toml into
// dir if one doesn't already exist there. It reports whether it
// created one; a false, nil result means config.toml already exists
// and nothing was done.
func BootstrapDefaultConfig(dir string) (bool, error) {
	path := filepath.Join(dir, "config.toml")
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false, err
	}
	if err := os.WriteFile(path, defaultConfigTOML, 0o644); err != nil {
		return false, err
	}
	return true, nil
}
