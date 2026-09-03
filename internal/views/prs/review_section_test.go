package prs

import (
	tea "charm.land/bubbletea/v2"

	"testing"
	"time"
)

func mkPR(url, title string, age time.Duration) pr {
	p := pr{Title: title, URL: url, UpdatedAt: time.Now().Add(-age)}
	p.Repository.NameWithOwner = "acme/repo"
	p.Number = 1
	return p
}

func TestApplySortBuildsReviewSection(t *testing.T) {
	v := &View{showReview: true}
	v.list.SetRowHeight(2)
	v.raw = []pr{mkPR("u1", "mine-old", 2*time.Hour), mkPR("u2", "mine-new", time.Minute)}
	v.reviewRaw = []pr{mkPR("u3", "theirs", time.Hour)}
	v.applySort()

	if got := v.list.Total(); got != 4 {
		t.Fatalf("Total = %d, want 2 mine + separator + 1 review", got)
	}
	// Separator sits between the sections and review PRs follow it.
	if !v.list.Any(func(p pr) bool { return p.Separator != "" }) {
		t.Error("no separator row in the combined list")
	}
	if v.list.Selected().Title != "mine-new" {
		t.Errorf("first selection = %q, want the newest own PR", v.list.Selected().Title)
	}

	// Toggled off, the review section disappears entirely.
	v.showReview = false
	v.applySort()
	if got := v.list.Total(); got != 2 {
		t.Errorf("Total after toggle = %d, want 2", got)
	}

	// No review PRs: no dangling separator.
	v.showReview, v.reviewRaw = true, nil
	v.applySort()
	if v.list.Any(func(p pr) bool { return p.Separator != "" }) {
		t.Error("separator rendered with an empty review section")
	}
}

func TestNotifyNewReviewsGating(t *testing.T) {
	v := &View{}
	if cmd := v.notifyNewReviews(nil, []pr{mkPR("u1", "t", 0)}); cmd != nil {
		t.Error("no notifier: expected nil command")
	}
	v.notifier = fakeNotifier{}
	if cmd := v.notifyNewReviews(nil, []pr{mkPR("u1", "t", 0)}); cmd != nil {
		t.Error("unseeded: expected nil command")
	}
	v.seeded = true
	if cmd := v.notifyNewReviews([]pr{mkPR("u1", "t", 0)}, []pr{mkPR("u1", "t", 0), mkPR("u2", "t2", 0)}); cmd == nil {
		t.Error("new review request: expected a command")
	}
	if cmd := v.notifyNewReviews([]pr{mkPR("u1", "t", 0)}, []pr{mkPR("u1", "t", 0)}); cmd != nil {
		t.Error("unchanged set: expected nil command")
	}
}

type fakeNotifier struct{}

func (fakeNotifier) Notify(title, body string) tea.Msg { return nil }
