package campagne

import "testing"

// TestProdDepsDownPollsServerHop — le câblage du sens download : shaper ET
// compteurs sur l'émission serveur (veth-s via netns cgo-srv). Sonde veth-c
// en download = compter les ACK (gel 0,0 malgré des Mo reçus — prouvé sur
// le banc : serveur 24 Mo envoyés, gel 0,0). Sens montant inchangé.
func TestProdDepsDownPollsServerHop(t *testing.T) {
	up := ProdDeps()
	if up.ShaperIf != "veth-c" || up.Direction != "" {
		t.Fatalf("up: ShaperIf=%q Direction=%q", up.ShaperIf, up.Direction)
	}
	if up.StatsFn == nil {
		t.Fatal("up: StatsFn doit exister (compteurs émission cliente)")
	}
	down := ProdDepsDown()
	if down.ShaperIf != "veth-s" || down.Direction != "down" {
		t.Fatalf("down: ShaperIf=%q Direction=%q", down.ShaperIf, down.Direction)
	}
	if down.StatsFn == nil {
		t.Fatal("down: StatsFn doit exister (compteurs émission serveur)")
	}
}
