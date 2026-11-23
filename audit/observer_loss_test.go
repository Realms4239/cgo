package audit

import (
	"math"
	"strings"
	"testing"
)

// ObserverLoss — budget de perte d'observateur : le compteur réel de réservations
// manquées sur un dénominateur réel d'événements tentés. Aucun nombre fabriqué.
func TestObserverLossFromRealCounters(t *testing.T) {
	events, pct, denom := ObserverLoss(1000, 3)
	if events != 1000 {
		t.Fatalf("events = %d, voulu 1000", events)
	}
	if math.Abs(pct-0.3) > 1e-9 {
		t.Fatalf("pct = %v, voulu 0.3", pct)
	}
	if !strings.Contains(denom, "retrans_events") || !strings.Contains(denom, "qdisc_drop_packets") {
		t.Fatalf("dénominateur %q ne nomme pas les compteurs réels", denom)
	}
}

// Oracle hostile : zéro réservation manquée donne 0 % mais le dénominateur
// reste nommé — on ne peut pas lire « pas de donnée » à la place de « 0 % ».
func TestObserverLossZeroFailuresStatesDenominator(t *testing.T) {
	events, pct, denom := ObserverLoss(1000, 0)
	if events != 1000 || pct != 0 {
		t.Fatalf("events/pct = %d/%v, voulu 1000/0", events, pct)
	}
	if denom == "" {
		t.Fatal("zéro perte doit quand même nommer le dénominateur")
	}
}

// Sans événement observé, on n'invente ni compteur ni pourcentage ; le
// dénominateur est toujours déclaré.
func TestObserverLossNoEventsFabricatesNothing(t *testing.T) {
	events, pct, denom := ObserverLoss(0, 0)
	if events != 0 || pct != 0 {
		t.Fatalf("events/pct = %d/%v, voulu 0/0", events, pct)
	}
	if denom == "" {
		t.Fatal("dénominateur requis même sans événement")
	}
}
