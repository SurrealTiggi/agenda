package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	d := Default()
	if len(d.Views) != 3 {
		t.Errorf("default Views = %v, want 3 entries", d.Views)
	}
	if d.GitHub.Filter == "" {
		t.Error("default GitHub.Filter is empty")
	}
	if !d.SessionsEnabled() {
		t.Error("sessions should be enabled by default")
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // empty dir: no config.yml
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil for missing file", err)
	}
	if cfg.GitHub.Filter != Default().GitHub.Filter {
		t.Errorf("GitHub.Filter = %q, want default", cfg.GitHub.Filter)
	}
}

func TestLoadMergesOntoDefaults(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	dir := filepath.Join(xdg, "agenda")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Only override a couple of fields; the rest must keep their defaults.
	yaml := "github:\n  filter: \"org:acme review-requested:@me\"\nlinear:\n  token: lin_api_secret\nsessions:\n  enabled: false\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.GitHub.Filter != "org:acme review-requested:@me" {
		t.Errorf("GitHub.Filter = %q, want overridden value", cfg.GitHub.Filter)
	}
	if cfg.Linear.Token != "lin_api_secret" {
		t.Errorf("Linear.Token = %q, want overridden value", cfg.Linear.Token)
	}
	if cfg.SessionsEnabled() {
		t.Error("sessions enabled, want disabled by file")
	}
	// Views weren't in the file, so they should still be the default.
	if len(cfg.Views) != 3 {
		t.Errorf("Views = %v, want default 3 entries (not clobbered)", cfg.Views)
	}
}

func TestPathHonorsXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
	got, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if want := "/custom/xdg/agenda/config.yml"; got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}
