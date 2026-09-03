package ui

import "testing"

func TestResolvePalette(t *testing.T) {
	// Empty name is the terminal-ANSI default.
	p, err := ResolvePalette("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Accent != "13" {
		t.Errorf("default Accent = %q, want ANSI 13", p.Accent)
	}

	// A named theme with a manual override on top.
	p, err = ResolvePalette("catppuccin-mocha", map[string]string{"accent": "#f5c2e7"})
	if err != nil {
		t.Fatal(err)
	}
	if p.Accent != "#f5c2e7" {
		t.Errorf("Accent = %q, want the override", p.Accent)
	}
	if p.Green != "#a6e3a1" {
		t.Errorf("Green = %q, want the mocha green (not clobbered)", p.Green)
	}

	if _, err := ResolvePalette("solarized", nil); err == nil {
		t.Error("unknown theme name should error")
	}
	if _, err := ResolvePalette("nord", map[string]string{"pink": "#fff"}); err == nil {
		t.Error("unknown palette key should error")
	}
}

func TestPaletteNamesStable(t *testing.T) {
	names := PaletteNames()
	if len(names) != len(builtins) {
		t.Fatalf("PaletteNames() has %d entries, want %d", len(names), len(builtins))
	}
	if names[0] != "default" {
		t.Errorf("names[0] = %q, want default first", names[0])
	}
	for _, n := range names {
		if _, ok := builtins[n]; !ok {
			t.Errorf("name %q not in builtins", n)
		}
	}
}
