# agenda

A terminal dashboard that unifies the things you keep checking into one TUI you
tab between — your open **GitHub PRs**, your local **agent sessions** (Claude
Code, Codex, Antigravity), and your **Linear issues**. Each is a distinct
*view*; switch with `tab` / `shift+tab`.

Built with [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) and
[Lip Gloss](https://github.com/charmbracelet/lipgloss). Inspired by
[gh-dash](https://github.com/dlvhdr/gh-dash).

## Install

```sh
go install github.com/obliadp/agenda@latest
```

Requirements:
- A **Nerd Font** in your terminal (for the status glyphs).
- The **`gh` CLI**, authenticated (`gh auth login`) — powers the PRs view.

## Configuration

Config lives at `$XDG_CONFIG_HOME/agenda/config.yml` (defaults to
`~/.config/agenda/config.yml`). It's optional — agenda runs with sensible
defaults. See [`config.example.yml`](./config.example.yml) for all options.

The only view needing setup is **Linear**: add a personal API key
(linear.app → Settings → Security & access → API keys):

```yaml
linear:
  token: lin_api_xxx
```

## Keys

Global: `tab` / `shift+tab` switch views · `ctrl+r` refresh · `q` quit · `/` filter

| View     | Keys |
|----------|------|
| PRs      | `enter` open · `d` diff · `y` copy URL |
| Sessions | `enter` resume · `s` cycle sort |
| Linear   | `enter` open · `y` copy URL · `b` copy branch |

## Views

- **PRs** — fetched via `gh api graphql`, showing CI/review status, diff size,
  comments, and labels, with a glamour-rendered description.
- **Sessions** — scans `~/.claude`, `~/.codex`, and `~/.gemini/antigravity-cli`;
  `enter` resumes the selected session in its original directory.
- **Linear** — issues assigned to you, via the Linear GraphQL API.
