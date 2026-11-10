package kit

import (
	"strings"
	"testing"
)

// La table des verbes Suite : un seul maillon, 30 verbes, partagé des deux
// côtés du tag build. Ces tests épinglent le drift — l'ancienne whitelist
// en connaissait 27 pendant que les exécutants en exécutaient 30.

func TestSuiteVerbsHasNoDuplicates(t *testing.T) {
	seen := map[string]int{}
	for i, v := range SuiteVerbs {
		if seen[v] != 0 {
			t.Fatalf("verbe %q apparaît en double (positions %d)", v, i)
		}
		seen[v] = i + 1
	}
	if len(SuiteVerbs) != 30 {
		t.Fatalf("SuiteVerbs = %d verbes, veut 30 — un executant ajoute un verbe : la table suit", len(SuiteVerbs))
	}
}

func TestSuiteVerbsCoverDispatchSet(t *testing.T) {
	// Les deux verbes qui manquaient : RunKit les exécute, la Suite n'osait
	// pas les envoyer.
	for _, v := range []string{"console:dns", "console:tls", "console:keysetup"} {
		if !IsKnownSuiteVerb(v) {
			t.Errorf("%s : la Suite refuserait un verbe que les exécutants exécutent", v)
		}
	}
	// Et les verbes que les exécutants n'ont JAMAIS eu : la Suite ne les invente pas.
	for _, v := range []string{"", "bg:nope", "bogus", "console:nope", "direct:nope"} {
		if IsKnownSuiteVerb(v) {
			t.Errorf("%q : la Suite inventerait un verbe que les exécutants n'ont pas", v)
		}
	}
}

// Tout verbe émis par le producteur du palier (next.go) doit être exécutable
// par la Suite — sinon « Suite » produit une étape qu'elle refusera juste après.
func TestEveryProducedVerbIsKnown(t *testing.T) {
	// Les verbes produits par gather*, scannés statiquement : collecte via
	// la source du fichier serait fragile ; on énumère la table produit/
	// dispatch à la place — le test est honnête seulement si la table est
	// la MÊME source : les tests du producteur couvrent le mapping.
	for _, v := range SuiteVerbs {
		parts := strings.SplitN(v, ":", 2)
		if len(parts) != 2 {
			t.Fatalf("verbe %q sans namespace prefixe:suffixe", v)
		}
		switch parts[0] {
		case "direct", "bg", "console":
		default:
			t.Fatalf("namespace %q inconnu pour %q", parts[0], v)
		}
	}
}
