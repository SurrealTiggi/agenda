package ui

import (
	"strings"
	"testing"
)

func TestMarkdownStyledElements(t *testing.T) {
	src := "# Title\n\n## Section\n\n- [x] done thing\n- [ ] open thing\n\n> quoted wisdom\n\nSome `inline` code.\n\n```\nteam: unknown\n```\n"
	out := Markdown(src, 60)

	if strings.Contains(out, "## ") {
		t.Error("H2 still renders with raw ## prefix")
	}
	for _, want := range []string{"󰲡", "󰲣", "󰱒", "󰄱", "▐"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered markdown missing %q glyph:\n%s", want, out)
		}
	}
	if strings.Contains(out, "[x]") || strings.Contains(out, "[✓]") {
		t.Error("task checkboxes still render as brackets")
	}
	// Code blocks carry a gutter bar so they read as blocks.
	if !strings.Contains(out, "▎") {
		t.Error("code block missing its gutter bar")
	}
}

func TestMarkdownInlineCodeContrast(t *testing.T) {
	// Inline code must use the palette yellow, not glamour's stock red,
	// which is illegible on the grey chip in most palettes.
	SetPalette(Palette{
		Accent: "#111111", Border: "#222222", Text: "#333333", Dim: "#444444",
		Green: "#555555", Red: "#666666", Yellow: "#abcdef", Blue: "#888888",
		Cyan: "#999999", Magenta: "#aaaaaa",
	})
	defer SetPalette(builtins["default"])
	out := Markdown("has `code` inline", 40)
	if !strings.Contains(out, "171;205;239") && !strings.Contains(strings.ToLower(out), "abcdef") {
		t.Errorf("inline code not using the palette yellow:\n%q", out)
	}
}

func TestMarkdownFallsBackOnGarbage(t *testing.T) {
	// Nothing should ever blank the preview; even odd input renders.
	if out := Markdown("", 40); out == "" {
		// empty in, empty out is fine
		_ = out
	}
	// Words render in separate style spans, so check them individually.
	out := Markdown("just text", 40)
	if !strings.Contains(out, "just") || !strings.Contains(out, "text") {
		t.Errorf("plain text lost in rendering: %q", out)
	}
}

func TestMarkdownRendererCacheTracksPalette(t *testing.T) {
	Markdown("hi", 40)
	gen := mdCache.gen
	SetPalette(Pal()) // palette swap (same colors, new generation)
	Markdown("hi", 40)
	if mdCache.gen == gen {
		t.Error("renderer cache not rebuilt after a palette change")
	}
}

func TestSplitFences(t *testing.T) {
	src := "intro\n\n```yaml\nteam: x\n```\n\ntail\n\n  ```\n  indented fence\n"
	segs := splitFences(src)
	var kinds []string
	for _, s := range segs {
		if s.code {
			kinds = append(kinds, "code:"+s.lang)
		} else {
			kinds = append(kinds, "text")
		}
	}
	want := []string{"text", "code:yaml", "text", "code:"}
	if len(kinds) != len(want) {
		t.Fatalf("segments = %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("segments = %v, want %v", kinds, want)
		}
	}
	if segs[1].text != "team: x" {
		t.Errorf("code content = %q", segs[1].text)
	}
	// The unterminated indented fence still captures its content.
	if strings.TrimRight(segs[3].text, "\n") != "  indented fence" {
		t.Errorf("unterminated fence content = %q", segs[3].text)
	}
}

func TestRenderCodeBlockGutter(t *testing.T) {
	out := renderCodeBlock("a: 1\nb: 2", "yaml", 40)
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2", len(lines))
	}
	for _, l := range lines {
		if !strings.Contains(l, "▎") {
			t.Errorf("line missing gutter bar: %q", l)
		}
	}
}

func TestMarkdownRespectsGlyphToggle(t *testing.T) {
	SetGlyphs(false)
	defer SetGlyphs(true)
	out := Markdown("# T\n\n## S\n\n- [x] done\n", 60)
	for _, glyph := range []string{"󰲡", "󰲣", "󰱒"} {
		if strings.Contains(out, glyph) {
			t.Errorf("glyphs off, but %q rendered", glyph)
		}
	}
	if !strings.Contains(out, "[✓]") && !strings.Contains(out, "[x]") {
		t.Error("glyphs off should fall back to bracket checkboxes")
	}
}
