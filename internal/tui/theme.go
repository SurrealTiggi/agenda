package tui

import "charm.land/lipgloss/v2"

// theme holds the styles shared across the chrome. Views may define their own
// row styling but should pull accent colors from here for consistency.
type theme struct {
	tabActive   lipgloss.Style
	tabInactive lipgloss.Style
	tabBar      lipgloss.Style
	preview     lipgloss.Style
	footer      lipgloss.Style
	footerKey   lipgloss.Style
	footerDesc  lipgloss.Style
	footerSep   lipgloss.Style
}

func defaultTheme() theme {
	var (
		accent = lipgloss.Color("13") // magenta
		dim    = lipgloss.Color("8")  // bright black / grey
		border = lipgloss.Color("8")
	)
	return theme{
		tabActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(accent).
			Bold(true).
			Padding(0, 2),
		tabInactive: lipgloss.NewStyle().
			Foreground(dim).
			Padding(0, 2),
		tabBar: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(border).
			BorderBottom(true).
			BorderTop(false).BorderLeft(false).BorderRight(false),
		preview: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(border).
			BorderLeft(true).
			BorderTop(false).BorderBottom(false).BorderRight(false).
			PaddingLeft(2),
		footer: lipgloss.NewStyle().
			Foreground(dim),
		footerKey: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),
		footerDesc: lipgloss.NewStyle().
			Foreground(dim),
		footerSep: lipgloss.NewStyle().
			Foreground(dim).
			SetString(" · "),
	}
}
