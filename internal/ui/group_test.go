package ui

import (
	"strings"
	"testing"
	"time"
)

func TestTimeBucketAt(t *testing.T) {
	now := time.Date(2026, 9, 1, 15, 0, 0, 0, time.Local)
	cases := map[string]string{}
	stamp := func(d time.Duration) time.Time { return now.Add(-d) }
	_ = cases

	tests := []struct {
		at   time.Time
		want string
	}{
		{stamp(time.Hour), "Today"},
		{time.Date(2026, 9, 1, 0, 0, 1, 0, time.Local), "Today"},
		{time.Date(2026, 8, 31, 23, 0, 0, 0, time.Local), "Yesterday"},
		{time.Date(2026, 8, 31, 1, 0, 0, 0, time.Local), "Yesterday"},
		{stamp(3 * 24 * time.Hour), "Last 7 days"},
		{stamp(20 * 24 * time.Hour), "Last 30 days"},
		{stamp(90 * 24 * time.Hour), "Older"},
	}
	for _, tc := range tests {
		if got := TimeBucketAt(tc.at, now); got != tc.want {
			t.Errorf("TimeBucketAt(%v) = %q, want %q", tc.at, got, tc.want)
		}
	}
}

func TestInsertGroups(t *testing.T) {
	items := []string{"a1", "a2", "b1", "c1", "c2"}
	label := func(s string) string { return s[:1] }
	header := func(l string) string { return "H:" + l }
	got := InsertGroups(items, label, header)
	want := []string{"H:a", "a1", "a2", "H:b", "b1", "H:c", "c1", "c2"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	if out := InsertGroups(nil, label, header); len(out) != 0 {
		t.Errorf("empty in, got %v", out)
	}
}

func TestGroupHeaderShape(t *testing.T) {
	h := GroupHeader("In Progress", 40)
	if !strings.HasPrefix(h, "\n") {
		t.Error("header should start with a blank line (two-line row layout)")
	}
	if !strings.Contains(h, "In Progress") || !strings.Contains(h, "─") {
		t.Errorf("header missing label or rule: %q", h)
	}
}
