package sessions

import (
	"fmt"
	"strings"
)

// rate is the per-1M-token USD price for one model family. cacheRead and
// cacheWrite are derived from in at construction (0.125x and 1.25x), matching
// Anthropic's billing structure.
type rate struct{ in, out, cacheRead, cacheWrite float64 }

func newRate(in, out float64) rate {
	return rate{in: in, out: out, cacheRead: in * 0.125, cacheWrite: in * 1.25}
}

// rateFor maps a log's model string to a rate by family substring. Fable/Mythos
// is checked first (top of the price range) and is also the default, so an
// unrecognised model is never under-priced. Anthropic exposes no pricing API, so
// these are maintained by hand; update when list prices change.
func rateFor(model string) rate {
	m := strings.ToLower(model)
	switch {
	case strings.Contains(m, "fable"), strings.Contains(m, "mythos"):
		return newRate(10, 50)
	case strings.Contains(m, "opus"):
		return newRate(5, 25)
	case strings.Contains(m, "sonnet"):
		return newRate(3, 15)
	case strings.Contains(m, "haiku"):
		return newRate(1, 5)
	default:
		return newRate(10, 50) // conservative: highest tier
	}
}

// costUSD estimates the USD cost of a session from its token usage and model.
func costUSD(in, out, cacheRead, cacheWrite int, model string) float64 {
	r := rateFor(model)
	const perM = 1_000_000.0
	return float64(in)/perM*r.in +
		float64(out)/perM*r.out +
		float64(cacheRead)/perM*r.cacheRead +
		float64(cacheWrite)/perM*r.cacheWrite
}

// fmtCost renders a cost for display: "" for exactly zero, "<$0.01" for tiny
// nonzero amounts, else "$X.XX".
func fmtCost(c float64) string {
	switch {
	case c == 0:
		return ""
	case c < 0.01:
		return "<$0.01"
	default:
		return fmt.Sprintf("$%.2f", c)
	}
}
