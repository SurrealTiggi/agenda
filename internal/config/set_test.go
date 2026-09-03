package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetPreservesCommentsAndKeys(t *testing.T) {
	writeConfig(t, `# my precious comment
views: [prs, linear]

linear:
  # token comment
  token: lin_api_secret
  filter:
    include_completed: false
`)
	if err := Set("theme.name", "nord"); err != nil {
		t.Fatal(err)
	}
	if err := Set("linear.filter.include_completed", true); err != nil {
		t.Fatal(err)
	}

	path, _ := Path()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, want := range []string{"# my precious comment", "# token comment", "lin_api_secret", "name: nord", "include_completed: true"} {
		if !strings.Contains(s, want) {
			t.Errorf("config after Set missing %q:\n%s", want, s)
		}
	}

	// The result still loads, with both changes visible.
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Theme.Name != "nord" || !cfg.Linear.Filter.IncludeCompleted {
		t.Errorf("reloaded cfg = theme %q, include_completed %v", cfg.Theme.Name, cfg.Linear.Filter.IncludeCompleted)
	}
}

func TestSetCreatesMissingFile(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	if err := Set("refresh.every", "5m"); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.RefreshFor("prs"); got.String() != "5m0s" {
		t.Errorf("RefreshFor(prs) = %v, want 5m", got)
	}
	fi, err := os.Stat(filepath.Join(xdg, "agenda", "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("new config mode = %v, want 0600", fi.Mode().Perm())
	}
}

func TestSetReplacesScalarWithMap(t *testing.T) {
	writeConfig(t, "linear: \"\"\n")
	if err := Set("linear.token", "lin_api_x"); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Linear.Token != "lin_api_x" {
		t.Errorf("token = %q after replacing scalar parent", cfg.Linear.Token)
	}
}
