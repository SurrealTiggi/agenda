package prs

import (
	"strings"
	"testing"
	"time"
)

func intp(i int) *int { return &i }

func mkThread(id, path string, line *int, resolved, outdated bool, bodies ...string) prThread {
	t := prThread{ID: id, Path: path, Line: line, IsResolved: resolved, IsOutdated: outdated}
	for i, b := range bodies {
		c := prComment{Body: b, CreatedAt: time.Now().Add(time.Duration(i) * time.Minute)}
		c.Author.Login = "alice"
		t.Comments.Nodes = append(t.Comments.Nodes, c)
	}
	return t
}

func TestThreadAnnotations(t *testing.T) {
	anns := threadAnnotations([]prThread{
		mkThread("t1", "foo.go", intp(42), false, false, "why this?"),
		mkThread("t2", "bar.go", nil, true, true, "old note"),
	}, 80)

	if anns[0].Path != "foo.go" || anns[0].Line != 42 {
		t.Errorf("ann[0] pinned at %s:%d, want foo.go:42", anns[0].Path, anns[0].Line)
	}
	if !strings.Contains(anns[0].Text, "@alice") || !strings.Contains(anns[0].Text, "why this?") {
		t.Errorf("ann text missing author/body: %q", anns[0].Text)
	}
	// Outdated threads pin at the file header (line 0).
	if anns[1].Line != 0 {
		t.Errorf("outdated thread pinned at line %d, want 0", anns[1].Line)
	}
	if !strings.Contains(anns[1].Text, "resolved") {
		t.Errorf("resolved thread not marked: %q", anns[1].Text)
	}
}

func TestRenderCommentsPane(t *testing.T) {
	var data prComments
	c := prComment{Body: "first!", CreatedAt: time.Now().Add(-2 * time.Hour)}
	c.Author.Login = "bob"
	data.Comments.Nodes = []prComment{c}
	r := prReview{State: "APPROVED", Body: "ship it", CreatedAt: time.Now().Add(-time.Hour)}
	r.Author.Login = "carol"
	data.Reviews.Nodes = []prReview{r}
	// Empty COMMENTED shell reviews (containers for inline threads) are hidden.
	shell := prReview{State: "COMMENTED", CreatedAt: time.Now()}
	shell.Author.Login = "dave"
	data.Reviews.Nodes = append(data.Reviews.Nodes, shell)
	data.ReviewThreads.Nodes = []prThread{mkThread("t1", "foo.go", intp(3), false, false, "hm")}

	out, anchors := renderCommentsPane(data, 60)
	if !strings.Contains(out, "@bob") || !strings.Contains(out, "first") {
		t.Error("conversation comment missing")
	}
	if !strings.Contains(out, "@carol") || !strings.Contains(out, "approved") {
		t.Error("review verdict missing")
	}
	if strings.Contains(out, "@dave") {
		t.Error("empty COMMENTED shell review should be hidden")
	}
	if len(anchors) != 1 || anchors[0].ID != "t1" {
		t.Fatalf("anchors = %v, want one for t1", anchors)
	}
	// The anchor points at the thread's header line.
	lines := strings.Split(out, "\n")
	if !strings.Contains(lines[anchors[0].Line], "foo.go:3") {
		t.Errorf("anchor line = %q, want the thread header", lines[anchors[0].Line])
	}
}
