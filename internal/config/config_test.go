package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear any env vars that could interfere.
	t.Setenv("FM_CACHE_DIR", "")
	t.Setenv("FM_OUTPUT", "")

	cfg, err := Load("", "")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	home, _ := os.UserHomeDir()
	wantDir := filepath.Join(home, ".local/share/fm/cache")
	if cfg.CacheDir != wantDir {
		t.Errorf("CacheDir = %q, want %q", cfg.CacheDir, wantDir)
	}
	if cfg.OutputFormat != "text" {
		t.Errorf("OutputFormat = %q, want %q", cfg.OutputFormat, "text")
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("FM_CACHE_DIR", "/tmp/fm-cache")
	t.Setenv("FM_OUTPUT", "json")

	cfg, err := Load("", "")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.CacheDir != "/tmp/fm-cache" {
		t.Errorf("CacheDir = %q, want %q", cfg.CacheDir, "/tmp/fm-cache")
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("OutputFormat = %q, want %q", cfg.OutputFormat, "json")
	}
}

func TestLoadFlagOverridesEnv(t *testing.T) {
	t.Setenv("FM_CACHE_DIR", "/tmp/fm-cache-env")
	t.Setenv("FM_OUTPUT", "json")

	cfg, err := Load("/tmp/fm-cache-flag", "markdown")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.CacheDir != "/tmp/fm-cache-flag" {
		t.Errorf("CacheDir = %q, want %q", cfg.CacheDir, "/tmp/fm-cache-flag")
	}
	if cfg.OutputFormat != "markdown" {
		t.Errorf("OutputFormat = %q, want %q", cfg.OutputFormat, "markdown")
	}
}

func TestLoadTildeExpansion(t *testing.T) {
	t.Setenv("FM_CACHE_DIR", "~/my-fm-cache")
	t.Setenv("FM_OUTPUT", "")

	cfg, err := Load("", "")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	home, _ := os.UserHomeDir()
	want := filepath.Join(home, "my-fm-cache")
	if cfg.CacheDir != want {
		t.Errorf("CacheDir = %q, want %q", cfg.CacheDir, want)
	}
}

func TestLoadAbsolutePathNotModified(t *testing.T) {
	t.Setenv("FM_CACHE_DIR", "/absolute/path")
	t.Setenv("FM_OUTPUT", "")

	cfg, err := Load("", "")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.CacheDir != "/absolute/path" {
		t.Errorf("CacheDir = %q, want %q", cfg.CacheDir, "/absolute/path")
	}
}
