package campagne

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// topoEnv pointe CGO_TOPOLOGY_JSON sur un fichier de test — le chemin par
// défaut (/tmp/cgo-dualns-topology.json) n'est JAMAIS touché par les tests.
func topoEnv(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "topology.json")
	t.Setenv("CGO_TOPOLOGY_JSON", p)
	return p
}

func writeTopo(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestCheckBenchTopologyAbsentFileIsLegal(t *testing.T) {
	topoEnv(t) // chemin inexistant dans un répertoire temporaire
	if err := CheckBenchTopology("ingress"); err != nil {
		t.Fatalf("fichier absent doit rester légal : %v", err)
	}
	if err := CheckBenchTopology(""); err != nil {
		t.Fatalf("attendu vide ⇒ egress normalisé : %v", err)
	}
}

func TestCheckBenchTopologyMatch(t *testing.T) {
	p := topoEnv(t)
	writeTopo(t, p, `{"shape_mode":"ingress","rate_mbit":20}`)
	if err := CheckBenchTopology("ingress"); err != nil {
		t.Fatalf("accord ingress doit passer : %v", err)
	}
	writeTopo(t, p, `{"shape_mode":"egress"}`)
	if err := CheckBenchTopology("egress"); err != nil {
		t.Fatalf("accord egress doit passer : %v", err)
	}
	// expected vide ⇒ egress normalisé.
	if err := CheckBenchTopology(""); err != nil {
		t.Fatalf("accord egress (attendu vide) doit passer : %v", err)
	}
}

func TestCheckBenchTopologyMismatchRefused(t *testing.T) {
	p := topoEnv(t)
	writeTopo(t, p, `{"shape_mode":"ingress"}`)
	err := CheckBenchTopology("egress")
	if err == nil {
		t.Fatal("banc ingress + campagne egress doit être refusé")
	}
	if !strings.Contains(err.Error(), "ingress") || !strings.Contains(err.Error(), "egress") {
		t.Fatalf("l'erreur doit nommer les DEUX modes : %v", err)
	}
	writeTopo(t, p, `{"shape_mode":"egress"}`)
	if err := CheckBenchTopology("ingress"); err == nil {
		t.Fatal("banc egress + campagne ingress doit être refusé")
	}
}

func TestCheckBenchTopologyGarbageRefused(t *testing.T) {
	p := topoEnv(t)
	writeTopo(t, p, `{not json`)
	if err := CheckBenchTopology("egress"); err == nil {
		t.Fatal("JSON invalide doit être refusé (échec fermé)")
	}
}

func TestCheckBenchTopologyMissingShapeModeRefused(t *testing.T) {
	p := topoEnv(t)
	writeTopo(t, p, `{"rate_mbit":20}`)
	if err := CheckBenchTopology("egress"); err == nil {
		t.Fatal("shape_mode absent doit être refusé (échec fermé)")
	}
}
