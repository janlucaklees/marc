package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

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
