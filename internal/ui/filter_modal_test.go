package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func typeRunes(m *FilterModal, s string) {
	for _, r := range s {
		m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

func TestFilterModalTypingAndApply(t *testing.T) {
	m := NewFilterModal("Filter PRs", "", []string{"repo", "branch", "title"}, nil, false)
	typeRunes(&m, "oauth")
	if m.Query() != "oauth" {
		t.Fatalf("Query() = %q, want oauth", m.Query())
	}
	if done, cancelled := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter}); !done || cancelled {
		t.Errorf("enter: done=%v cancelled=%v, want true/false", done, cancelled)
	}
}

func TestFilterModalToggleFields(t *testing.T) {
	m := NewFilterModal("x", "", []string{"repo", "branch", "title"}, nil, false)
	// Switch focus to the field list, move to "branch", toggle it OFF.
	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown}) // repo -> branch
	m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	got := m.EnabledFields()
	// branch removed; repo and title remain.
	if strings.Join(got, ",") != "repo,title" {
		t.Errorf("EnabledFields() = %v, want [repo title]", got)
	}
}

func TestFilterModalCaseSensitiveRow(t *testing.T) {
	m := NewFilterModal("x", "", []string{"repo"}, nil, false)
	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})  // focus list, cursor on "repo"
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown}) // move onto the case-sensitive row
	m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	if !m.CaseSensitive() {
		t.Errorf("CaseSensitive() = false after toggling its row, want true")
	}
}

func TestFilterModalCancel(t *testing.T) {
	m := NewFilterModal("x", "", []string{"repo"}, nil, false)
	if done, cancelled := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape}); done || !cancelled {
		t.Errorf("esc: done=%v cancelled=%v, want false/true", done, cancelled)
	}
}

func TestFilterModalBackspace(t *testing.T) {
	m := NewFilterModal("x", "ab", []string{"repo"}, nil, false)
	m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if m.Query() != "a" {
		t.Errorf("Query() = %q after backspace, want a", m.Query())
	}
}
