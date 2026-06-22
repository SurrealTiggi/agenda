package ui

import "charm.land/lipgloss/v2"

// AgentIcon returns the colored Nerd Font glyph for an agent tool
// ("claude" | "codex" | "agy"), used wherever a session is shown so the agent
// is recognizable at a glance without spelling out its name.
func AgentIcon(tool string) string {
	glyph, color := IconAgentClaude, "5" // claude: magenta
	switch tool {
	case "codex":
		glyph, color = IconAgentCodex, "2" // green
	case "agy":
		glyph, color = IconAgentAgy, "4" // blue
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(glyph)
}
