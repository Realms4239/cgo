package audit

import "testing"

// Les bornes sont la règle du moteur : une mesure ne dure pas une journée,
// le régime de charge reste dans la fenêtre documentée. Avant, la borne de
// durée n'existait que chez le handler HTTP — le CLI passait 999999 s au
// moteur (27 h d'audit).

func TestClampDuration(t *testing.T) {
	cases := []struct{ in, want int }{
		{0, MinDurationSec},
		{-5, MinDurationSec},
		{2, 2}, // audit court : le moteur le permet
		{10, 10},
		{60, 60},
		{600, 600},
		{999999, MaxDurationSec},
	}
	for _, c := range cases {
		if got := ClampDuration(c.in); got != c.want {
			t.Errorf("ClampDuration(%d) = %d, veut %d", c.in, got, c.want)
		}
	}
}

// La fenêtre d'opérateur (10–600 s) est plus stricte que le garde-fou du
// moteur : c'est ce que l'API rabote et ce que le CLI refuse.
func TestClampWindow(t *testing.T) {
	cases := []struct{ in, want int }{
		{2, MinDurationSec},
		{10, 10},
		{60, 60},
		{999999, MaxDurationSec},
	}
	for _, c := range cases {
		if got := ClampWindow(c.in); got != c.want {
			t.Errorf("ClampWindow(%d) = %d, veut %d", c.in, got, c.want)
		}
	}
}

func TestClampFlows(t *testing.T) {
	cases := []struct{ in, want int }{
		{0, DefaultFlows},
		{-1, DefaultFlows},
		{1, 1},
		{4, 4},
		{8, 8},
		{9, MaxFlows},
		{64, MaxFlows},
	}
	for _, c := range cases {
		if got := ClampFlows(c.in); got != c.want {
			t.Errorf("ClampFlows(%d) = %d, veut %d", c.in, got, c.want)
		}
	}
}

func TestBoundsPredicates(t *testing.T) {
	if DurationInBounds(999999) {
		t.Error("999999 s serait dans les bornes")
	}
	if DurationInBounds(5) {
		t.Error("5 s serait dans les bornes")
	}
	if !DurationInBounds(60) {
		t.Error("60 s devrait être dans les bornes")
	}
	if FlowsInBounds(9) {
		t.Error("9 flux seraient dans les bornes")
	}
	if !FlowsInBounds(8) {
		t.Error("8 flux devraient être dans les bornes")
	}
}

// Le plan de phases est dérivé de la durée : une durée hors bornes ne doit
// pas produire un plan de 27 h.
func TestPhasePlanFollowsClampedDuration(t *testing.T) {
	idle, loaded, rec := phasePlan(ClampDuration(999999), true)
	total := idle + loaded + rec
	if total > MaxDurationSec {
		t.Fatalf("plan de %d s pour une durée clamped à %d s", total, MaxDurationSec)
	}
}
