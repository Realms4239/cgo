package qdisc

import (
	"strings"
	"testing"
)

// TestApplyNetemPhysicsLegacyUniform — la perte UNIFORME reste la forme
// historique exacte (ge nil) : le corpus gelé se reproduit par cet appel.
// (Remplace TestApplyNetemBurstLoss, dont l'ancienne forme à 5 arguments
// « p r h 1-r k » était invalide sur Linux et testait H=200 — une proba
// de 200. Voir ge.go pour la forme documentée à 4 arguments.)
func TestApplyNetemPhysicsLegacyUniform(t *testing.T) {
	f := &FakeRunner{}
	if err := ApplyNetemPhysics(f, "veth-c", 100, 15, 0.5, nil); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(f.Calls[0], " ")
	if !strings.Contains(joined, "qdisc replace dev veth-c root handle 1: netem delay 100ms 15ms loss 0.5%") {
		t.Fatalf("forme uniforme historique : %s", joined)
	}

	// perte nulle : pas de token « loss » du tout — profil P1 propre.
	f2 := &FakeRunner{}
	if err := ApplyNetemPhysics(f2, "veth-c", 20, 2, 0, nil); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(f2.Calls[0], " "), "loss") {
		t.Fatalf("P1 sans perte ne doit pas émettre « loss » : %s", strings.Join(f2.Calls[0], " "))
	}
}
