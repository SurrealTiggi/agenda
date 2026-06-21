package sessions

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// tool identifies which agent produced a session.
type tool string

const (
	toolClaude tool = "claude"
	toolCodex  tool = "codex"
	toolAgy    tool = "agy"
)

const maxTitle = 120

// meta is the parsed summary of one session file (the cacheable part).
type meta struct {
	Cwd       string `json:"cwd"`
	Title     string `json:"title"`
	Msgs      int    `json:"msgs"`
	SessionID string `json:"session_id"`
}

// session is one row: parsed meta plus filesystem facts.
type session struct {
	meta
	Tool  tool
	Path  string
	MTime time.Time
}

// --- discovery --------------------------------------------------------------

func home() string {
	h, _ := os.UserHomeDir()
	return h
}

func claudeDir() string { return filepath.Join(home(), ".claude", "projects") }
func codexDir() string  { return filepath.Join(home(), ".codex", "sessions") }
func agyDir() string    { return filepath.Join(home(), ".gemini", "antigravity-cli") }

func agyConvDir() string  { return filepath.Join(agyDir(), "conversations") }
func agyBrainDir() string { return filepath.Join(agyDir(), "brain") }
func agyLastConv() string { return filepath.Join(agyDir(), "cache", "last_conversations.json") }

func agyTranscript(convID string) string {
	return filepath.Join(agyBrainDir(), convID, ".system_generated", "logs", "transcript.jsonl")
}

// discover returns every session file with its tool tag.
func discover() []struct {
	path string
	tool tool
} {
	var out []struct {
		path string
		tool tool
	}
	add := func(p string, t tool) {
		out = append(out, struct {
			path string
			tool tool
		}{p, t})
	}

	// Claude: top-level *.jsonl per project (skip nested subagent logs).
	if entries, err := os.ReadDir(claudeDir()); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			matches, _ := filepath.Glob(filepath.Join(claudeDir(), e.Name(), "*.jsonl"))
			for _, m := range matches {
				add(m, toolClaude)
			}
		}
	}

	// Codex: rollout-*.jsonl anywhere under the sessions tree.
	_ = filepath.WalkDir(codexDir(), func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), "rollout-") && strings.HasSuffix(d.Name(), ".jsonl") {
			add(p, toolCodex)
		}
		return nil
	})

	// Antigravity: one .db per conversation.
	if matches, err := filepath.Glob(filepath.Join(agyConvDir(), "*.db")); err == nil {
		for _, m := range matches {
			add(m, toolAgy)
		}
	}

	return out
}

// --- text helpers -----------------------------------------------------------

var wsRe = regexp.MustCompile(`\s+`)

func cleanText(s string) string {
	return strings.TrimSpace(wsRe.ReplaceAllString(s, " "))
}

// textFromContent pulls plain text out of a Claude/Codex content field, which
// is either a JSON string or a list of typed blocks.
func textFromContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var parts []string
		for _, b := range blocks {
			if b.Type == "text" || b.Type == "input_text" {
				parts = append(parts, b.Text)
			}
		}
		return strings.Join(parts, " ")
	}
	return ""
}

// isRealUserText filters out tool results, system reminders, and caveats so
// the title reflects an actual human prompt.
func isRealUserText(s string) bool {
	s = strings.TrimSpace(s)
	switch {
	case s == "":
		return false
	case strings.HasPrefix(s, "<"):
		return false
	case strings.Contains(s, "tool_use_id"):
		return false
	case strings.HasPrefix(s, "Caveat:"):
		return false
	case strings.HasPrefix(s, "[Request interrupted"):
		return false
	}
	return true
}

func truncTitle(s string) string {
	r := []rune(s)
	if len(r) > maxTitle {
		return string(r[:maxTitle])
	}
	return s
}

// --- parsers ----------------------------------------------------------------

// scanLines runs fn over each JSON-decoded line of a file. Decode errors on a
// line are skipped (logs can contain partial trailing writes).
func scanLines(path string, fn func(line []byte)) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		fn(sc.Bytes())
	}
	return sc.Err()
}

func parseClaude(path string) meta {
	var cwd, first, last, aiTitle string
	n := 0
	_ = scanLines(path, func(line []byte) {
		var d struct {
			Type    string `json:"type"`
			Cwd     string `json:"cwd"`
			AiTitle string `json:"aiTitle"`
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(line, &d) != nil {
			return
		}
		if cwd == "" && d.Cwd != "" {
			cwd = d.Cwd
		}
		switch d.Type {
		case "ai-title":
			if d.AiTitle != "" {
				aiTitle = cleanText(d.AiTitle)
			}
		case "user":
			text := cleanText(textFromContent(d.Message.Content))
			if isRealUserText(text) {
				n++
				if first == "" {
					first = text
				}
				last = text
			}
		}
	})
	title := aiTitle
	if title == "" {
		title = last
	}
	if title == "" {
		title = first
	}
	return meta{Cwd: cwd, Title: truncTitle(title), Msgs: n, SessionID: stem(path)}
}

var uuidRe = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

func parseCodex(path string) meta {
	var cwd, sid, first, last string
	n := 0
	_ = scanLines(path, func(line []byte) {
		var d struct {
			Type    string `json:"type"`
			Payload struct {
				Cwd     string          `json:"cwd"`
				ID      string          `json:"id"`
				Type    string          `json:"type"`
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"payload"`
		}
		if json.Unmarshal(line, &d) != nil {
			return
		}
		switch {
		case d.Type == "session_meta":
			if d.Payload.Cwd != "" {
				cwd = d.Payload.Cwd
			}
			if d.Payload.ID != "" {
				sid = d.Payload.ID
			}
		case d.Type == "response_item" && d.Payload.Type == "message" && d.Payload.Role == "user":
			text := cleanText(textFromContent(d.Payload.Content))
			if isRealUserText(text) {
				n++
				if first == "" {
					first = text
				}
				last = text
			}
		}
	})
	if sid == "" {
		if m := uuidRe.FindString(filepath.Base(path)); m != "" {
			sid = m
		} else {
			sid = stem(path)
		}
	}
	title := first
	if title == "" {
		title = last
	}
	return meta{Cwd: cwd, Title: truncTitle(title), Msgs: n, SessionID: sid}
}

var (
	agyReqRe = regexp.MustCompile(`(?s)<USER_REQUEST>\s*(.*?)\s*</USER_REQUEST>`)
	agyWsRe  = regexp.MustCompile(`(?s)\[CorpusName\]:\s*\n([^\n]+?)\s*->`)
)

func agyUserText(content string) string {
	if m := agyReqRe.FindStringSubmatch(content); m != nil {
		return cleanText(m[1])
	}
	return ""
}

// agyCwdFromDB recovers the workspace dir from the '[URI] -> [CorpusName]'
// mapping Antigravity embeds in the conversation db.
func agyCwdFromDB(dbPath string) string {
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return ""
	}
	m := agyWsRe.FindSubmatch(data)
	if m == nil {
		return ""
	}
	uri := strings.TrimSpace(string(m[1]))
	uri = strings.TrimPrefix(uri, "file://")
	if strings.HasPrefix(uri, "/") {
		return uri
	}
	return ""
}

// agyCwdFallback maps conversation-id -> cwd from last_conversations.json.
func agyCwdFallback() map[string]string {
	raw, err := os.ReadFile(agyLastConv())
	if err != nil {
		return nil
	}
	var byCwd map[string]string // cwd -> id
	if json.Unmarshal(raw, &byCwd) != nil {
		return nil
	}
	out := make(map[string]string, len(byCwd))
	for cwd, id := range byCwd {
		out[id] = cwd
	}
	return out
}

func parseAgy(path string) meta {
	convID := stem(path)
	var first, last string
	n := 0
	if tp := agyTranscript(convID); fileExists(tp) {
		_ = scanLines(tp, func(line []byte) {
			var d struct {
				Type    string `json:"type"`
				Source  string `json:"source"`
				Content string `json:"content"`
			}
			if json.Unmarshal(line, &d) != nil {
				return
			}
			if d.Type == "USER_INPUT" && d.Source == "USER_EXPLICIT" {
				if text := agyUserText(d.Content); text != "" {
					n++
					if first == "" {
						first = text
					}
					last = text
				}
			}
		})
	}
	title := first
	if title == "" {
		title = last
	}
	return meta{Cwd: agyCwdFromDB(path), Title: truncTitle(title), Msgs: n, SessionID: convID}
}

// turn is one message in a session, for the preview pane.
type turn struct {
	role string // "user" or "assistant"
	text string
}

// conversationTurns returns the chronological user/assistant turns of a
// session, for previewing. Ported from the Python tool's _conversation_turns.
func conversationTurns(path string, t tool) []turn {
	var turns []turn
	switch t {
	case toolAgy:
		tp := agyTranscript(stem(path))
		if !fileExists(tp) {
			return nil
		}
		_ = scanLines(tp, func(line []byte) {
			var d struct {
				Type    string `json:"type"`
				Source  string `json:"source"`
				Content string `json:"content"`
			}
			if json.Unmarshal(line, &d) != nil {
				return
			}
			switch {
			case d.Type == "USER_INPUT" && d.Source == "USER_EXPLICIT":
				if txt := agyUserText(d.Content); txt != "" {
					turns = append(turns, turn{"user", txt})
				}
			case d.Type == "PLANNER_RESPONSE" && d.Source == "MODEL":
				if txt := cleanText(d.Content); txt != "" {
					turns = append(turns, turn{"assistant", txt})
				}
			}
		})

	case toolCodex:
		_ = scanLines(path, func(line []byte) {
			var d struct {
				Payload struct {
					Type    string          `json:"type"`
					Role    string          `json:"role"`
					Content json.RawMessage `json:"content"`
				} `json:"payload"`
			}
			if json.Unmarshal(line, &d) != nil || d.Payload.Type != "message" {
				return
			}
			role := d.Payload.Role
			if role != "user" && role != "assistant" {
				return
			}
			txt := cleanText(textFromContent(d.Payload.Content))
			if txt == "" || (role == "user" && !isRealUserText(txt)) {
				return
			}
			turns = append(turns, turn{role, txt})
		})

	default: // claude
		_ = scanLines(path, func(line []byte) {
			var d struct {
				Type    string `json:"type"`
				Message struct {
					Content json.RawMessage `json:"content"`
				} `json:"message"`
			}
			if json.Unmarshal(line, &d) != nil {
				return
			}
			if d.Type != "user" && d.Type != "assistant" {
				return
			}
			txt := cleanText(textFromContent(d.Message.Content))
			if txt == "" || (d.Type == "user" && !isRealUserText(txt)) {
				return
			}
			turns = append(turns, turn{d.Type, txt})
		})
	}
	return turns
}

func parse(path string, t tool) meta {
	switch t {
	case toolCodex:
		return parseCodex(path)
	case toolAgy:
		return parseAgy(path)
	default:
		return parseClaude(path)
	}
}

// --- small fs helpers -------------------------------------------------------

func stem(path string) string {
	b := filepath.Base(path)
	return strings.TrimSuffix(b, filepath.Ext(b))
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func shortenPath(p string) string {
	if p == "" {
		return "?"
	}
	h := home()
	if p == h {
		return "~"
	}
	if strings.HasPrefix(p, h+"/") {
		return "~" + strings.TrimPrefix(p, h)
	}
	return p
}
