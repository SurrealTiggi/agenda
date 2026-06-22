package prs

import "testing"

func TestLinearRef(t *testing.T) {
	cases := []struct {
		name string
		pr   pr
		want string
	}{
		{
			name: "uppercase in title parens",
			pr:   pr{Title: "docs: plan for shard migration (SRE-4419)"},
			want: "SRE-4419",
		},
		{
			name: "lowercase in title prefix",
			pr:   pr{Title: "ci(sre-4228): migrate to orb"},
			want: "SRE-4228",
		},
		{
			name: "from branch when title has none",
			pr:   pr{Title: "feat: add gateway routes", HeadRefName: "orjan/sre-3717-add-gateway"},
			want: "SRE-3717",
		},
		{
			name: "title wins over branch",
			pr:   pr{Title: "ENG-1 do thing", HeadRefName: "user/sre-2-other"},
			want: "ENG-1",
		},
		{
			name: "from body as last resort",
			pr:   pr{Title: "fix flaky test", Body: "Closes OPS-77 after the rollout."},
			want: "OPS-77",
		},
		{
			name: "no reference",
			pr:   pr{Title: "fix the bug", HeadRefName: "fix/the-bug"},
			want: "",
		},
		{
			name: "version token is not an issue ref",
			pr:   pr{Title: "bump to v2", HeadRefName: "chore/v2-bump"},
			want: "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.pr.linearRef(); got != c.want {
				t.Errorf("linearRef() = %q, want %q", got, c.want)
			}
		})
	}
}
