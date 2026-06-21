// Command agenda is a terminal dashboard that unifies several "views" — your
// open GitHub PRs, your local agent sessions, and your Linear issues — into a
// single TUI you tab between. Configuration (including any personal details
// like a Linear API token) lives in $XDG_CONFIG_HOME/agenda/config.yml.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/obliadp/agenda/internal/config"
	"github.com/obliadp/agenda/internal/tui"
	"github.com/obliadp/agenda/internal/views/linear"
	"github.com/obliadp/agenda/internal/views/prs"
	"github.com/obliadp/agenda/internal/views/sessions"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "agenda: config error:", err)
		os.Exit(1)
	}

	// Build the configured views in tab order. (Placeholders for now — each
	// will be swapped for its real implementation.)
	var views []tui.View
	for _, name := range cfg.Views {
		switch name {
		case "prs":
			views = append(views, prs.New(cfg.GitHub.Filter))
		case "sessions":
			if cfg.SessionsEnabled() {
				views = append(views, sessions.New())
			}
		case "linear":
			views = append(views, linear.New(cfg.Linear.Token))
		}
	}

	p := tea.NewProgram(tui.New(cfg, views))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "agenda:", err)
		os.Exit(1)
	}
}
