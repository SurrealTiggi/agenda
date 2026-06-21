package ui

import (
	"testing"
	"time"
)

func TestAge(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name string
		t    time.Time
		want string
	}{
		{"zero", time.Time{}, ""},
		{"minutes", now.Add(-5 * time.Minute), "5m"},
		{"hours", now.Add(-90 * time.Minute), "1h"},
		{"days", now.Add(-50 * time.Hour), "2d"},
		{"months", now.Add(-60 * 24 * time.Hour), "2mo"},
		{"future clamps to 0m", now.Add(10 * time.Minute), "0m"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Age(c.t); got != c.want {
				t.Errorf("Age(%v) = %q, want %q", c.t, got, c.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		in   string
		n    int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 3, "he…"},
		{"hello", 1, "…"},
		{"hello", 0, ""},
		{"hello", -2, ""},
		{"héllo", 3, "hé…"}, // multibyte counted by rune, not byte
	}
	for _, c := range cases {
		if got := Truncate(c.in, c.n); got != c.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", c.in, c.n, got, c.want)
		}
	}
}
