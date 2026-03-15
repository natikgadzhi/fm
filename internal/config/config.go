// Package config handles configuration loading from environment variables and defaults.
package config

import (
	"os"
	"path/filepath"
)

// Config holds the application configuration.
type Config struct {
	// CacheDir is the directory for cached email files.
	CacheDir string

	// OutputFormat is the default output format (text, json, markdown).
	OutputFormat string
}

// Load reads configuration from environment variables and applies defaults.
// Flag overrides (non-empty flagCacheDir / flagOutputFormat) take precedence.
func Load(flagCacheDir, flagOutputFormat string) (*Config, error) {
	cfg := &Config{
		CacheDir:     defaultCacheDir(),
		OutputFormat: "text",
	}

	// Environment variables override defaults.
	if v := os.Getenv("FM_CACHE_DIR"); v != "" {
		cfg.CacheDir = v
	}
	if v := os.Getenv("FM_OUTPUT"); v != "" {
		cfg.OutputFormat = v
	}

	// CLI flags override environment variables.
	if flagCacheDir != "" {
		cfg.CacheDir = flagCacheDir
	}
	if flagOutputFormat != "" {
		cfg.OutputFormat = flagOutputFormat
	}

	// Expand ~ in CacheDir.
	cfg.CacheDir = expandHome(cfg.CacheDir)

	return cfg, nil
}

// defaultCacheDir returns the default cache directory path.
func defaultCacheDir() string {
	return "~/.local/share/fm/cache"
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) string {
	if len(path) == 0 {
		return path
	}
	if path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}
