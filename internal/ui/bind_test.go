package ui

import "testing"

func TestBind(t *testing.T) {
	b := Bind([]string{"q", "ctrl+c"}, "", "quit")
	if got := b.Keys(); len(got) != 2 || got[0] != "q" {
		t.Errorf("Keys() = %v, want [q ctrl+c]", got)
	}
	if h := b.Help(); h.Key != "q" || h.Desc != "quit" {
		t.Errorf("Help() = %+v, want derived key q", h)
	}

	// helpKey override wins over derivation.
	if h := Bind([]string{"shift+tab"}, "custom", "prev").Help(); h.Key != "custom" {
		t.Errorf("Help().Key = %q, want the explicit override", h.Key)
	}

	// The compact glyph forms are derived for modifier keys.
	if h := Bind([]string{"shift+up"}, "", "scroll").Help(); h.Key != "⇧↑" {
		t.Errorf("Help().Key = %q, want ⇧↑", h.Key)
	}

	// No keys at all disables the binding entirely.
	if b := Bind(nil, "", "gone"); b.Enabled() {
		t.Error("Bind(nil) should be disabled")
	}
}
