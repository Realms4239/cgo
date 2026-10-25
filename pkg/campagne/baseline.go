package campagne

import (
	"math"

	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/model"
)

// BaselineStable decides G6: is the idle baseline quiet enough that a loaded
// measurement taken after it can be trusted?
//
// THE PROBLEM WITH A STATIC THRESHOLD
// The original rule was `p95 - median < max(5 ms, 20% of median)`. That is a
// sensible guard for a wired or fibre link, where an idle baseline should be
// almost perfectly flat. It is wrong for a LEO satellite link, where the
// physical layer itself moves: every inter-satellite handoff shifts the path
// length, so the idle RTT distribution is legitimately wide. On the frozen
// corpus every P4 (Starlink) row fails G6 — 42 of 42, spread 17-19.5 ms against
// a threshold of about 8 ms — and not one of them is a measurement fault. The
// gate was measuring orbital mechanics and calling it instability.
//
// WHAT THE THRESHOLD SHOULD BE
// The gate exists to catch a noisy *host*, not a noisy *link*. The link's own
// jitter is a declared property of the profile (Profile.JitterMs), so the
// allowance must be at least that large: a baseline spread within the profile's
// declared jitter is the link behaving as specified. Above that, the excess is
// what the host added.
//
// So the allowance is:
//
//	floor + jitterAllowance * declaredJitter
//
// with the floor kept from the original rule (a wired link still gets its
// absolute and relative allowance) and the jitter term added on top. A profile
// that declares no jitter therefore behaves exactly as before, which keeps the
// rule backwards-compatible for P1/P2/P3.
//
// jitterAllowance is 2.0 rather than 1.0 because a p95-minus-median spread is
// not the same statistic as a declared peak-to-peak jitter: for a roughly
// symmetric distribution the p95-median gap is about half the p5-p95 range, so
// a link whose declared jitter is J can legitimately show a spread near J
// without anything being wrong. 2.0 covers that with margin and still fails a
// host that is genuinely thrashing.
const jitterAllowance = 2.0

// BaselineStable reports whether the idle baseline passes G6, and the threshold
// it was judged against so the decision is auditable in the event record.
//
// Returns (ok, thresholdMs). A baseline with too few samples fails: G6 needs
// enough points for a p95 to mean anything.
func BaselineStable(baseRTT []float64, prof model.Profile) (bool, float64) {
	if len(baseRTT) <= 4 {
		// Not enough samples to judge stability at all. The threshold is still
		// reported so the operator can see what would have applied.
		return false, baselineThreshold(0, prof)
	}
	sum := metrics.Summarize(baseRTT)
	thr := baselineThreshold(sum.Median, prof)
	return sum.P95-sum.Median < thr, thr
}

// baselineThreshold is the G6 allowance in milliseconds for a given baseline
// median and profile.
func baselineThreshold(medianMs float64, prof model.Profile) float64 {
	// The original host-noise floor, unchanged: an absolute 5 ms plus a
	// relative 20% of the median, whichever is larger.
	floor := math.Max(5, 0.2*medianMs)
	// The link's own declared jitter, which the host cannot be blamed for.
	link := jitterAllowance * prof.JitterMs
	return floor + link
}
