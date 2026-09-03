package ui

import (
	"strconv"
	"strings"
)

// maxDiffLines caps rendered diffs so a giant PR can't stall the UI; the tail
// says how much was cut.
const maxDiffLines = 4000

// DiffAnnotation is an inline note pinned into a rendered diff: a review
// thread on a file/line. Line is the right-side (new file) line number; 0
// pins the note right under the file header. Text is the pre-rendered block.
type DiffAnnotation struct {
	ID   string // thread id, echoed back in the anchor
	Path string
	Line int
	Text string
}

// DiffAnchor reports where an annotation landed in the rendered output, so a
// caller can jump the viewport between annotations.
type DiffAnchor struct {
	Line int // 0-based line index into the rendered output
	ID   string
}

// RenderDiff colorizes a unified diff for the preview pane. See
// RenderAnnotatedDiff for the annotated variant.
func RenderDiff(text string, width int) string {
	out, _ := RenderAnnotatedDiff(text, width, nil)
	return out
}

// RenderAnnotatedDiff colorizes a unified diff and pins annotations under
// the lines they reference: file headers stand out, hunk headers are cyan,
// additions green, deletions red, context dimmed. Annotations whose line no
// longer exists in the diff are appended at the end under their file:line
// label rather than dropped.
func RenderAnnotatedDiff(text string, width int, anns []DiffAnnotation) (string, []DiffAnchor) {
	src := strings.Split(strings.TrimRight(text, "\n"), "\n")
	truncated := false
	if len(src) > maxDiffLines {
		src, truncated = src[:maxDiffLines], true
	}

	placed := make([]bool, len(anns))
	var out []string
	var anchors []DiffAnchor

	emit := func(i int) {
		anchors = append(anchors, DiffAnchor{Line: len(out), ID: anns[i].ID})
		out = append(out, strings.Split(anns[i].Text, "\n")...)
		placed[i] = true
	}

	curPath := ""
	newLine := 0 // right-side line number of the next content line
	for _, line := range src {
		switch {
		case strings.HasPrefix(line, "diff --git"):
			curPath = line
			if i := strings.LastIndex(line, " b/"); i >= 0 {
				curPath = line[i+3:]
			}
			out = append(out, "", Accent.Bold(true).Render(Truncate("▸ "+curPath, max(1, width))))
			for i, a := range anns {
				if !placed[i] && a.Path == curPath && a.Line == 0 {
					emit(i)
				}
			}
			continue
		case strings.HasPrefix(line, "@@"):
			// "@@ -a,b +c,d @@": the next right-side content line is c.
			newLine = hunkNewStart(line)
			out = append(out, Cyan.Render(Truncate(expandTabs(line), max(1, width))))
			continue
		}

		out = append(out, renderDiffLine(line, width))
		if isRightSideLine(line) {
			for i, a := range anns {
				if !placed[i] && a.Path == curPath && a.Line == newLine {
					emit(i)
				}
			}
			newLine++
		}
	}

	// Whatever couldn't be pinned (outdated threads, truncated hunks) lands
	// in a trailing section instead of disappearing.
	first := true
	for i, a := range anns {
		if placed[i] {
			continue
		}
		if first {
			out = append(out, "", Dim.Render("── comments on outdated lines "+strings.Repeat("─", 10)))
			first = false
		}
		label := a.Path
		if a.Line > 0 {
			label += ":" + strconv.Itoa(a.Line)
		}
		out = append(out, Dim.Render(label))
		anchors = append(anchors, DiffAnchor{Line: len(out) - 1, ID: a.ID})
		out = append(out, strings.Split(a.Text, "\n")...)
	}

	if truncated {
		out = append(out, "", Faint.Render("… diff truncated"))
	}
	return strings.Join(out, "\n"), anchors
}

// isRightSideLine reports whether a diff body line occupies a line in the
// new file (context or addition).
func isRightSideLine(line string) bool {
	if line == "" {
		return true // empty context line
	}
	switch line[0] {
	case ' ':
		return true
	case '+':
		return !strings.HasPrefix(line, "+++ ")
	default:
		return false
	}
}

// hunkNewStart parses the right-side start line out of "@@ -a,b +c,d @@".
func hunkNewStart(line string) int {
	i := strings.Index(line, "+")
	if i < 0 {
		return 0
	}
	rest := line[i+1:]
	end := strings.IndexAny(rest, ", @")
	if end < 0 {
		end = len(rest)
	}
	n, err := strconv.Atoi(rest[:end])
	if err != nil {
		return 0
	}
	return n
}

func expandTabs(s string) string { return strings.ReplaceAll(s, "\t", "    ") }

func renderDiffLine(line string, width int) string {
	line = expandTabs(line)
	switch {
	case strings.HasPrefix(line, "index "),
		strings.HasPrefix(line, "new file"),
		strings.HasPrefix(line, "deleted file"),
		strings.HasPrefix(line, "similarity index"),
		strings.HasPrefix(line, "rename "),
		strings.HasPrefix(line, "old mode"),
		strings.HasPrefix(line, "new mode"),
		strings.HasPrefix(line, "Binary files"),
		strings.HasPrefix(line, "+++ "),
		strings.HasPrefix(line, "--- "):
		return Dim.Render(Truncate(line, max(1, width)))
	case strings.HasPrefix(line, "+"):
		return Green.Render(Truncate(line, max(1, width)))
	case strings.HasPrefix(line, "-"):
		return Red.Render(Truncate(line, max(1, width)))
	default:
		return Text.Render(Truncate(line, max(1, width)))
	}
}
