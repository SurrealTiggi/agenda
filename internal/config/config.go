// Package config loads agenda's user configuration from an XDG-compliant
// location. The tool ships with sensible defaults so it runs out of the box;
// personal details (a Linear API token, custom search filters) live in the
// config file rather than in code, keeping the binary generic and shareable.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the fully-resolved configuration (defaults merged with the file).
type Config struct {
	// Views lists which views to show, in tab order. Recognised names:
	// "prs", "sessions", "linear".
	Views []string `yaml:"views"`

	GitHub   GitHubConfig   `yaml:"github"`
	Linear   LinearConfig   `yaml:"linear"`
	Sessions SessionsConfig `yaml:"sessions"`
}

type GitHubConfig struct {
	// Filter is the search query for the PRs view, in `gh search prs` syntax.
	Filter string `yaml:"filter"`
}

type LinearConfig struct {
	// Token is a Linear personal API key (lin_api_...). Required for the
	// Linear view; when empty the view renders a setup hint instead.
	Token string `yaml:"token"`
	// Filter is an optional extra GraphQL issue-filter clause; when empty the
	// view shows issues assigned to the authenticated user.
	Filter string `yaml:"filter"`
}

type SessionsConfig struct {
	// Enabled toggles the sessions view. Defaults to true.
	Enabled *bool `yaml:"enabled"`
}

// Default returns the built-in configuration used when no file exists or to
// fill gaps in a partial file.
func Default() Config {
	enabled := true
	return Config{
		Views: []string{"prs", "sessions", "linear"},
		GitHub: GitHubConfig{
			Filter: "author:@me is:open archived:false",
		},
		Sessions: SessionsConfig{Enabled: &enabled},
	}
}

// Dir is the directory agenda reads its config from:
// $XDG_CONFIG_HOME/agenda, falling back to ~/.config/agenda.
func Dir() (string, error) {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "agenda"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "agenda"), nil
}

// Path is the full path to the config file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yml"), nil
}

// Load reads the config file, merging it onto the defaults. A missing file is
// not an error — the defaults are returned and the file path is reported so a
// caller can offer to scaffold one.
func Load() (Config, error) {
	cfg := Default()

	path, err := Path()
	if err != nil {
		return cfg, err
	}

	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("reading %s: %w", path, err)
	}

	// Unmarshal onto the defaults so absent keys keep their default value.
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing %s: %w", path, err)
	}
	return cfg, nil
}

// SessionsEnabled reports whether the sessions view is on (default true).
func (c Config) SessionsEnabled() bool {
	return c.Sessions.Enabled == nil || *c.Sessions.Enabled
}
