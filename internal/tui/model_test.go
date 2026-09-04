package tui

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/obliadp/agenda/internal/config"
)

// stubView is the minimal View for chrome-level tests.
type stubView struct{ title string }

func (s *stubView) Title() string           { return s.title }
func (s *stubView) Init() tea.Cmd           { return nil }
func (s *stubView) Update(tea.Msg) tea.Cmd  { return nil }
func (s *stubView) SetSize(int, int, int)   {}
func (s *stubView) ListView() string        { return "" }
func (s *stubView) PreviewView() string     { return "" }
func (s *stubView) Bindings() []key.Binding { return nil }
func (s *stubView) Status() string          { return "" }
func (s *stubView) InputActive() bool       { return false }
func (s *stubView) PreviewKey() string      { return "" }
func (s *stubView) Loading() bool           { return false }

func TestRefreshIntervalsMatchViewsByTitle(t *testing.T) {
	cfg := config.Default()
	every := config.Duration(5 * time.Minute)
	linear := config.Duration(90 * time.Second)
	cfg.Refresh = config.RefreshConfig{Every: every, Linear: &linear}

	views := []View{&stubView{"PRs"}, &stubView{"Sessions"}, &stubView{"Linear"}}
	got := refreshIntervals(cfg, views)
	want := []time.Duration{5 * time.Minute, 5 * time.Minute, 90 * time.Second}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("interval[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestViewIndexForKey(t *testing.T) {
	cases := map[string]int{
		"1":  0, // "1" jumps to the first view
		"2":  1,
		"5":  4, // mid-range exercises the s[0]-'1' formula
		"9":  8,
		"0":  -1, // 0 is not a view hotkey
		"a":  -1,
		"":   -1,
		"12": -1, // multi-char (e.g. a key name) is not a digit jump
	}
	for in, want := range cases {
		if got := viewIndexForKey(in); got != want {
			t.Errorf("viewIndexForKey(%q) = %d, want %d", in, got, want)
		}
	}
}
