package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// sepItem is a strItem that can also be a non-selectable separator.
type sepItem struct {
	text string
	sep  bool
}

func (s sepItem) Render(int, bool, Highlighter) string { return s.text }
func (s sepItem) Fields() []Field {
	if s.sep {
		return nil
	}
	return []Field{{Name: "text", Text: s.text}}
}
func (s sepItem) Filter() string   { return s.text }
func (s sepItem) Selectable() bool { return !s.sep }

func sepList(items ...sepItem) List[sepItem] {
	l := NewList[sepItem]()
	l.SetItems(items)
	l.SetSize(40, 10)
	return l
}

func TestSeparatorNavigationSkips(t *testing.T) {
	l := sepList(
		sepItem{text: "a"},
		sepItem{text: "-- sep --", sep: true},
		sepItem{text: "b"},
	)
	l.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	if got := l.Selected().text; got != "b" {
		t.Errorf("down over separator selected %q, want b", got)
	}
	l.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	if got := l.Selected().text; got != "a" {
		t.Errorf("up over separator selected %q, want a", got)
	}
}

func TestSeparatorNeverSelectedAtEdges(t *testing.T) {
	l := sepList(
		sepItem{text: "-- top --", sep: true},
		sepItem{text: "a"},
		sepItem{text: "b"},
		sepItem{text: "-- bottom --", sep: true},
	)
	if got := l.Selected().text; got != "a" {
		t.Errorf("initial selection = %q, want a (skipping leading separator)", got)
	}
	l.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	if got := l.Selected().text; got != "b" {
		t.Errorf("G selected %q, want b (skipping trailing separator)", got)
	}
	l.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	if got := l.Selected().text; got != "a" {
		t.Errorf("g selected %q, want a", got)
	}
}

func TestSeparatorDroppedWhileFiltering(t *testing.T) {
	l := sepList(
		sepItem{text: "alpha"},
		sepItem{text: "-- sep --", sep: true},
		sepItem{text: "beta"},
	)
	l.SetQuery("e") // matches "sep" label too if separators weren't dropped
	if l.Len() != 1 {
		t.Fatalf("filtered Len = %d, want 1 (beta only, no separator)", l.Len())
	}
	if got := l.Selected().text; got != "beta" {
		t.Errorf("filtered selection = %q, want beta", got)
	}
}

func TestGroupedFieldNamesSkipHeaders(t *testing.T) {
	l := sepList(
		sepItem{text: "-- lane --", sep: true},
		sepItem{text: "a"},
	)
	names := l.FieldNames()
	if len(names) != 1 || names[0] != "text" {
		t.Errorf("FieldNames() = %v, want [text] from the first real item", names)
	}
}

func TestJumpKeepsHeaderVisible(t *testing.T) {
	items := []sepItem{{text: "-- top --", sep: true}}
	for i := 0; i < 12; i++ {
		items = append(items, sepItem{text: string(rune('a' + i))})
	}
	l := sepList(items...)
	l.SetSize(40, 5) // window smaller than the list

	l.Update(tea.KeyPressMsg{Code: 'G', Text: "G"}) // bottom
	l.Update(tea.KeyPressMsg{Code: 'g', Text: "g"}) // back to top
	if got := l.Selected().text; got != "a" {
		t.Fatalf("g selected %q, want a", got)
	}
	if l.offset != 0 {
		t.Errorf("offset = %d after jump to top, want 0 so the lane header shows", l.offset)
	}
}
