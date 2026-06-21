// Package placeholder is a stub View used to validate the skeleton. It will be
// replaced by the real prs, sessions, and linear views.
package placeholder

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	title        string
	listW, prevW int
	height       int
}

func New(title string) *Model { return &Model{title: title} }

func (m *Model) Title() string              { return m.title }
func (m *Model) Init() tea.Cmd              { return nil }
func (m *Model) Update(msg tea.Msg) tea.Cmd { return nil }
func (m *Model) SetSize(listW, prevW, h int) {
	m.listW, m.prevW, m.height = listW, prevW, h
}
func (m *Model) ListView() string {
	return fmt.Sprintf("%s list\n(%d×%d)", m.title, m.listW, m.height)
}
func (m *Model) PreviewView() string {
	return fmt.Sprintf("%s preview\n(%d×%d)", m.title, m.prevW, m.height)
}
func (m *Model) Bindings() []key.Binding { return nil }
func (m *Model) Status() string          { return m.title }
func (m *Model) InputActive() bool       { return false }
func (m *Model) PreviewKey() string      { return m.title }
