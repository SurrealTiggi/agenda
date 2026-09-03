package ui

import "charm.land/bubbles/v2/key"

// Bind builds a key.Binding from a configured key list. helpKey overrides the
// footer's key label; when empty it is derived from the first key. An empty
// keys list disables the binding entirely (it never matches and shows no
// help), which is how a user turns an action off in config.
func Bind(keys []string, helpKey, desc string) key.Binding {
	if len(keys) == 0 {
		return key.NewBinding(key.WithDisabled())
	}
	if desc == "" {
		return key.NewBinding(key.WithKeys(keys...))
	}
	if helpKey == "" {
		helpKey = HelpLabel(keys[0])
	}
	return key.NewBinding(key.WithKeys(keys...), key.WithHelp(helpKey, desc))
}

// helpLabels maps key names to the compact glyphs the footer shows.
var helpLabels = map[string]string{
	"shift+up":   "⇧↑",
	"shift+down": "⇧↓",
	"shift+tab":  "⇧tab",
	"pgdown":     "pgdn",
	"up":         "↑",
	"down":       "↓",
	"left":       "←",
	"right":      "→",
}

// HelpLabel is the footer display form of a key name.
func HelpLabel(k string) string {
	if l, ok := helpLabels[k]; ok {
		return l
	}
	return k
}

// KeyResolver looks up the configured keys for an action, falling back to the
// given defaults. It decouples ui from the config package.
type KeyResolver func(action string, def ...string) []string
