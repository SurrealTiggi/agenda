package linear

import "testing"

func TestParsePRRef(t *testing.T) {
	cases := []struct {
		url       string
		wantOK    bool
		wantID    string
		wantLabel string
	}{
		{
			url:       "https://github.com/sanity-io/argocd-apps/pull/7314",
			wantOK:    true,
			wantID:    "https://github.com/sanity-io/argocd-apps/pull/7314",
			wantLabel: "PR  sanity-io/argocd-apps#7314",
		},
		{url: "https://github.com/sanity-io/so/issues/42", wantOK: false},
		{url: "https://linear.app/acme/issue/SRE-1", wantOK: false},
		{url: "", wantOK: false},
	}
	for _, c := range cases {
		ref, ok := parsePRRef(c.url)
		if ok != c.wantOK {
			t.Errorf("parsePRRef(%q) ok = %v, want %v", c.url, ok, c.wantOK)
			continue
		}
		if ok {
			if ref.Kind != "pr" || ref.ID != c.wantID || ref.Label != c.wantLabel {
				t.Errorf("parsePRRef(%q) = %+v, want kind=pr id=%q label=%q",
					c.url, ref, c.wantID, c.wantLabel)
			}
		}
	}
}
