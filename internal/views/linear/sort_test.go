package linear

import (
	"testing"
	"time"
)

// mk builds a minimal issue for ordering tests.
func mk(id, stateName, stateType string, priority int, ageMinutes int) issue {
	var i issue
	i.Identifier = id
	i.State.Name = stateName
	i.State.Type = stateType
	i.Priority = priority
	i.UpdatedAt = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC).
		Add(-time.Duration(ageMinutes) * time.Minute)
	return i
}

func ids(in []issue) []string {
	out := make([]string, len(in))
	for i, it := range in {
		out[i] = it.Identifier
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSortIssuesByDate(t *testing.T) {
	in := []issue{
		mk("A-1", "Todo", "unstarted", 0, 50),
		mk("A-2", "Todo", "unstarted", 0, 5),
		mk("A-3", "Todo", "unstarted", 0, 20),
	}
	want := []string{"A-2", "A-3", "A-1"} // newest first
	if got := ids(sortIssues(in, sortRecent, false)); !equal(got, want) {
		t.Errorf("sortRecent = %v, want %v", got, want)
	}
}

func TestSortIssuesByStatus(t *testing.T) {
	in := []issue{
		mk("B-1", "Backlog", "backlog", 1, 1),
		mk("B-2", "Todo", "unstarted", 0, 1),
		mk("B-3", "In Progress", "started", 4, 1),
		mk("B-4", "In Review", "started", 2, 1),
	}
	// started before unstarted before backlog; within "started", states group by
	// name ("In Progress" < "In Review") regardless of priority.
	want := []string{"B-3", "B-4", "B-2", "B-1"}
	if got := ids(sortIssues(in, sortStatus, false)); !equal(got, want) {
		t.Errorf("sortStatus = %v, want %v", got, want)
	}
}

func TestSortIssuesByStatusTiebreaks(t *testing.T) {
	in := []issue{
		mk("C-1", "Todo", "unstarted", 0, 1),  // no priority -> last
		mk("C-2", "Todo", "unstarted", 3, 1),  // medium
		mk("C-3", "Todo", "unstarted", 1, 1),  // urgent
		mk("C-4", "Todo", "unstarted", 3, 99), // medium, but older
	}
	want := []string{"C-3", "C-2", "C-4", "C-1"}
	if got := ids(sortIssues(in, sortStatus, false)); !equal(got, want) {
		t.Errorf("sortStatus tiebreaks = %v, want %v", got, want)
	}
}

func TestSortIssuesDoesNotMutateInput(t *testing.T) {
	in := []issue{
		mk("D-1", "Backlog", "backlog", 0, 1),
		mk("D-2", "In Progress", "started", 0, 99),
	}
	before := ids(in)
	sortIssues(in, sortStatus, false)
	if got := ids(in); !equal(got, before) {
		t.Errorf("input reordered: %v, want %v", got, before)
	}
}

func TestSortOrderCyclesAllModes(t *testing.T) {
	for _, m := range sortOrder {
		if sortName[m] == "" {
			t.Errorf("sort mode %d has no name", m)
		}
	}
}

// Reversing negates the comparison, so for inputs with no fully-tied items the
// reversed list is the forward list backwards, in every mode.
func TestSortIssuesReversedIsExactInverse(t *testing.T) {
	in := []issue{
		mk("E-1", "In Progress", "started", 1, 50),
		mk("E-2", "Backlog", "backlog", 0, 5),
		mk("E-3", "Todo", "unstarted", 3, 20),
		mk("E-4", "In Review", "started", 4, 7),
	}
	for _, mode := range sortOrder {
		fwd := ids(sortIssues(in, mode, false))
		rev := ids(sortIssues(in, mode, true))
		want := make([]string, len(fwd))
		for i, id := range fwd {
			want[len(fwd)-1-i] = id
		}
		if !equal(rev, want) {
			t.Errorf("%s: reversed = %v, want %v (inverse of %v)",
				sortName[mode], rev, want, fwd)
		}
	}
}
