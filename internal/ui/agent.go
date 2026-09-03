package ui

// AgentIcon returns the colored Nerd Font glyph for an agent tool
// ("claude" | "codex" | "agy"), used wherever a session is shown so the agent
// is recognizable at a glance without spelling out its name.
func AgentIcon(tool string) string {
	glyph, style := IconAgentClaude, Magenta
	switch tool {
	case "codex":
		glyph, style = IconAgentCodex, Green
	case "agy":
		glyph, style = IconAgentAgy, Blue
	}
	return style.Render(glyph)
}
