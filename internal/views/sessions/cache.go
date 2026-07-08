package sessions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// cacheVersion is bumped whenever meta's schema changes, so stale on-disk
// caches are discarded rather than read back missing new fields.
const cacheVersion = "v6"

// cacheEntry stores a parsed meta keyed by a cheap file signature so unchanged
// files are never re-parsed. This mirrors the Python tool's meta-cache.
type cacheEntry struct {
	Sig  string `json:"sig"`
	Meta meta   `json:"meta"`
}

func cacheFile() string {
	dir := os.Getenv("XDG_CACHE_HOME")
	if dir == "" {
		dir = filepath.Join(home(), ".cache")
	}
	return filepath.Join(dir, "agenda", "sessions-cache.json")
}

func loadCache() map[string]cacheEntry {
	raw, err := os.ReadFile(cacheFile())
	if err != nil {
		return map[string]cacheEntry{}
	}
	var c map[string]cacheEntry
	if json.Unmarshal(raw, &c) != nil {
		return map[string]cacheEntry{}
	}
	return c
}

func saveCache(c map[string]cacheEntry) {
	path := cacheFile()
	if os.MkdirAll(filepath.Dir(path), 0o755) != nil {
		return
	}
	if raw, err := json.Marshal(c); err == nil {
		_ = os.WriteFile(path, raw, 0o644)
	}
}

// collect scans every session, parsing only files whose signature changed
// since the last run, and returns them sorted newest-first.
func collect() []session {
	cache := loadCache()
	next := make(map[string]cacheEntry)
	files := discover()

	var agyFallback map[string]string
	for _, f := range files {
		if f.tool == toolAgy {
			agyFallback = agyCwdFallback()
			break
		}
	}

	out := make([]session, 0, len(files))
	for _, f := range files {
		st, err := os.Stat(f.path)
		if err != nil {
			continue
		}
		// The version prefix invalidates the whole cache when meta's schema
		// changes (e.g. when Mentions were added), forcing a re-parse.
		sig := fmt.Sprintf("%s:%d:%d", cacheVersion, st.ModTime().Unix(), st.Size())

		var m meta
		if c, ok := cache[f.path]; ok && c.Sig == sig {
			m = c.Meta
		} else {
			m = parse(f.path, f.tool)
			m.Mentions, m.Body = scanMentionsAndBody(f.path, f.tool)
		}
		next[f.path] = cacheEntry{Sig: sig, Meta: m}

		if f.tool == toolAgy && m.Cwd == "" {
			m.Cwd = agyFallback[m.SessionID]
		}

		out = append(out, session{
			meta:    m,
			Tool:    f.tool,
			Path:    f.path,
			Updated: updatedAt(m, st.ModTime()),
		})
	}

	saveCache(next)
	sort.Slice(out, func(i, j int) bool { return out[i].Updated.After(out[j].Updated) })
	markSpawned(out)
	return out
}

// markSpawned sets Spawned on any session whose log references the session ID of
// an agent (programmatic) session — i.e. it spawned that sub-session. Most agent
// sessions have no discoverable parent (hook/command-spawned with no
// back-reference), so this marks only the few that are genuinely linkable.
func markSpawned(sessions []session) {
	// Collect agent session IDs (the potential children).
	agentIDs := make([]string, 0)
	for _, s := range sessions {
		if s.isAgent() && s.SessionID != "" {
			agentIDs = append(agentIDs, s.SessionID)
		}
	}
	if len(agentIDs) == 0 {
		return
	}
	for i := range sessions {
		p := &sessions[i]
		if p.isAgent() || p.Tool != toolClaude {
			continue // only human Claude sessions spawn agent sub-sessions
		}
		data, err := os.ReadFile(p.Path)
		if err != nil {
			continue
		}
		blob := string(data)
		n := 0
		for _, id := range agentIDs {
			if id != p.SessionID && strings.Contains(blob, id) {
				n++
			}
		}
		p.Spawned = n
	}
}
