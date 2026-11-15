package probe

import "testing"

// TestSetSocketTokenContract — le crochet posé par l'hôte (la campagne
// l'installe à l'init pour traduire bbr3→bbr sur hôte provisionné). On
// vérifie le contrat de la seam : débrayage du défaut identitaire, dernier
// crochet gagne, puis restauration. Aucun réseau : seul le pointeur de
// fonction est observé.
func TestSetSocketTokenContract(t *testing.T) {
	old := socketToken
	t.Cleanup(func() { socketToken = old })

	if got := socketToken("bbr3"); got != "bbr3" {
		t.Fatalf("défaut identitaire attendu, socketToken(bbr3) = %q", got)
	}
	SetSocketToken(func(cc string) string {
		if cc == "bbr3" {
			return "bbr"
		}
		return cc
	})
	if got := socketToken("bbr3"); got != "bbr" {
		t.Fatalf("traduction posée, socketToken(bbr3) = %q, want « bbr »", got)
	}
	if got := socketToken("cubic"); got != "cubic" {
		t.Fatalf("les autres jetons traversent intacts, got %q", got)
	}
}
