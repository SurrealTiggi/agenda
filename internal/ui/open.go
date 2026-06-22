package ui

import (
	"os/exec"
	"runtime"

	tea "charm.land/bubbletea/v2"
)

// OpenURL returns a command that opens u in the default browser. A "" url is a
// no-op (returns a nil command).
func OpenURL(u string) tea.Cmd {
	if u == "" {
		return nil
	}
	return func() tea.Msg {
		var c *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			c = exec.Command("open", u)
		case "windows":
			c = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
		default:
			c = exec.Command("xdg-open", u)
		}
		_ = c.Start()
		return nil
	}
}
