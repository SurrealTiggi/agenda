package sessions

import (
	"math"
	"testing"
)

func TestRateForFamily(t *testing.T) {
	cases := map[string]float64{ // model substring -> expected input rate
		"claude-fable-5":    10,
		"claude-mythos-5":   10,
		"claude-opus-4-7":   5,
		"claude-opus-4-8":   5,
		"claude-sonnet-5":   3,
		"claude-haiku-4-5":  1,
		"something-unknown": 10, // default -> highest tier (Fable)
		"":                  10,
	}
	for model, want := range cases {
		if got := rateFor(model).in; got != want {
			t.Errorf("rateFor(%q).in = %v, want %v", model, got, want)
		}
	}
}

func TestRateForCacheMultipliers(t *testing.T) {
	// cache-read = 0.125x input, cache-write = 1.25x input, for every tier.
	for _, model := range []string{"opus", "sonnet", "haiku", "fable"} {
		r := rateFor(model)
		if math.Abs(r.cacheRead-r.in*0.125) > 1e-9 {
			t.Errorf("%s cacheRead = %v, want %v", model, r.cacheRead, r.in*0.125)
		}
		if math.Abs(r.cacheWrite-r.in*1.25) > 1e-9 {
			t.Errorf("%s cacheWrite = %v, want %v", model, r.cacheWrite, r.in*1.25)
		}
	}
}

func TestCostUSD(t *testing.T) {
	// Opus 4.x: 1M input, 1M output, 1M cache-read, 1M cache-write.
	// = 5 + 25 + 0.625 + 6.25 = 36.875
	got := costUSD(1_000_000, 1_000_000, 1_000_000, 1_000_000, "claude-opus-4-8")
	if math.Abs(got-36.875) > 1e-6 {
		t.Errorf("costUSD = %v, want 36.875", got)
	}
	if costUSD(0, 0, 0, 0, "claude-opus-4-8") != 0 {
		t.Errorf("zero tokens should cost 0")
	}
}

func TestFmtCost(t *testing.T) {
	cases := map[float64]string{
		0:     "",       // omit exact zero
		0.004: "<$0.01", // tiny nonzero
		0.04:  "$0.04",
		1.235: "$1.24", // rounds
		12.3:  "$12.30",
	}
	for in, want := range cases {
		if got := fmtCost(in); got != want {
			t.Errorf("fmtCost(%v) = %q, want %q", in, got, want)
		}
	}
}
