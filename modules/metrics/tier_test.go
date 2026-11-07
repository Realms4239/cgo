package metrics

import "testing"

// TestSetDefaultTier — le palier actif se choisit (pas seulement se liste) :
// inconnu = refus, connu = appliqué au calcul horaire. Verrouillé par mutex
// (lectures concurrentes des campagnes + écriture POST /api/cost/tier).
func TestSetDefaultTier(t *testing.T) {
	if err := SetDefaultTier("nope-inexistant"); err == nil {
		t.Fatal("palier inconnu accepté")
	}
	if err := SetDefaultTier("yas-ftth-100gb"); err != nil {
		t.Fatal(err)
	}
	if got := ActiveTierName(); got != "yas-ftth-100gb" {
		t.Fatalf("palier actif = %q", got)
	}
	// fibre ~10× moins chère que le mobile au Go : le même gaspillage doit
	// coûter moins cher en FTTH qu'en cellular mensuel
	fiber := CostARPerH(100 * 1024 * 1024)
	if err := SetDefaultTier("yas-month-4.5gb"); err != nil {
		t.Fatal(err)
	}
	mobile := CostARPerH(100 * 1024 * 1024)
	if !(fiber < mobile) {
		t.Fatalf("FTTH %v doit coûter moins que cellular %v", fiber, mobile)
	}
	if got := ActiveTierName(); got != "yas-month-4.5gb" {
		t.Fatalf("restauration défaut: %q", got)
	}
}
