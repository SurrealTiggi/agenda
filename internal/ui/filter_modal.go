package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// FilterModal is the gitui-style filter popup: a query text box, a list of
// toggleable fields, and a case-sensitivity toggle, in three boxed sections.
// The host model owns it, routes keys to Update while open, and composites
// View over its content (like Picker).
type FilterModal struct {
	title         string
	query         string
	fields        []fieldToggle
	caseSensitive bool

	focusList bool // false = query box focused, true = toggle list focused
	cursor    int  // index into the toggle rows: 0..len(fields)-1 = fields,
	// len(fields) = the case-sensitive row.
}

type fieldToggle struct {
	name string
	on   bool
}

// NewFilterModal builds the modal. fields is every field name in order;
// enabled is the subset currently on (empty means all on).
func NewFilterModal(title, query string, fields, enabled []string, caseSensitive bool) FilterModal {
	on := map[string]bool{}
	for _, n := range enabled {
		on[n] = true
	}
	allOn := len(enabled) == 0
	toggles := make([]fieldToggle, len(fields))
	for i, n := range fields {
		toggles[i] = fieldToggle{name: n, on: allOn || on[n]}
	}
	return FilterModal{title: title, query: query, fields: toggles, caseSensitive: caseSensitive}
}

// rowCount is the number of selectable toggle rows (fields + case row).
func (m *FilterModal) rowCount() int { return len(m.fields) + 1 }

// Update handles keys. done is true when the user applied (read state via
// Query/EnabledFields/CaseSensitive); cancelled is true when they dismissed it.
func (m *FilterModal) Update(msg tea.Msg) (done, cancelled bool) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return false, false
	}
	switch km.String() {
	case "enter":
		return true, false
	case "esc", "ctrl+c":
		return false, true
	case "tab":
		m.focusList = !m.focusList
		return false, false
	}

	if m.focusList {
		switch km.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < m.rowCount()-1 {
				m.cursor++
			}
		case " ", "space":
			m.toggle()
		}
		return false, false
	}

	// Query box focused.
	switch km.String() {
	case "backspace":
		if m.query != "" {
			r := []rune(m.query)
			m.query = string(r[:len(r)-1])
		}
	default:
		if s := km.String(); len(s) == 1 {
			m.query += s
		}
	}
	return false, false
}

// toggle flips the row under the cursor.
func (m *FilterModal) toggle() {
	if m.cursor == len(m.fields) {
		m.caseSensitive = !m.caseSensitive
		return
	}
	m.fields[m.cursor].on = !m.fields[m.cursor].on
}

func (m *FilterModal) Query() string       { return m.query }
func (m *FilterModal) CaseSensitive() bool { return m.caseSensitive }

// EnabledFields returns the field names toggled on, in order.
func (m *FilterModal) EnabledFields() []string {
	var out []string
	for _, f := range m.fields {
		if f.on {
			out = append(out, f.name)
		}
	}
	return out
}

// View renders the three-section boxed modal. The caller composites it over
// its content.
func (m *FilterModal) View() string {
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	faint := lipgloss.NewStyle().Faint(true)

	const innerW = 30
	divider := faint.Render(strings.Repeat("─", innerW))

	// Section 1: query box.
	qCursor := ""
	if !m.focusList {
		qCursor = "█"
	}
	query := m.query + qCursor

	// Section 2: field toggles.
	var fieldLines []string
	for i, f := range m.fields {
		fieldLines = append(fieldLines, m.toggleRow(f.name, f.on, m.focusList && m.cursor == i, accent))
	}

	// Section 3: case-sensitive toggle.
	caseRow := m.toggleRow("case sensitive", m.caseSensitive,
		m.focusList && m.cursor == len(m.fields), accent)

	body := strings.Join([]string{
		query,
		divider,
		strings.Join(fieldLines, "\n"),
		divider,
		caseRow,
		"",
		faint.Render("tab focus · space toggle · enter apply · esc cancel"),
	}, "\n")

	titled := lipgloss.NewStyle().Bold(true).Render(m.title) + "\n" + body

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("13")).
		Padding(1, 2).
		Render(titled)
}

// toggleRow renders one "[x] name" / "[ ] name" row with a cursor accent.
func (m *FilterModal) toggleRow(name string, on, cursor bool, accent lipgloss.Style) string {
	box := "[ ]"
	if on {
		box = "[x]"
	}
	bar := "  "
	if cursor {
		bar = accent.Render("▌") + " "
	}
	return bar + box + " " + name
}
