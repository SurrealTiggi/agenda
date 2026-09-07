package main

import "testing"

func TestVersionStringPrefersLdflags(t *testing.T) {
	defer func() { version, commit, date = "", "", "" }()

	version, commit, date = "1.2.3", "abc1234", "2026-09-07"
	if got, want := versionString(), "1.2.3 (abc1234, 2026-09-07)"; got != want {
		t.Errorf("versionString() = %q, want %q", got, want)
	}

	version, commit, date = "1.2.3", "", ""
	if got := versionString(); got != "1.2.3" {
		t.Errorf("versionString() = %q, want bare version without commit/date", got)
	}
}

func TestVersionStringFallsBackToBuildInfo(t *testing.T) {
	// Under `go test` there are no ldflags, so this exercises the
	// debug.ReadBuildInfo path; the exact value depends on the build,
	// but it must never be empty.
	if versionString() == "" {
		t.Error("versionString() = empty, want a devel/module fallback")
	}
}
