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

func TestCheckBenchTopologyAbsentFileFailsClosed(t *testing.T) {
	topoEnv(t) // chemin inexistant dans un répertoire temporaire
	if err := CheckBenchTopology("ingress"); err == nil {
		t.Fatal("missing bench topology must fail closed")
	}
	if err := CheckBenchTopology(""); err == nil {
		t.Fatal("missing bench topology must fail closed for the default direction too")
	}
}

func TestCheckBenchTopologyMatch(t *testing.T) {
	p := topoEnv(t)
	writeTopo(t, p, `{"shape_mode":"ingress","rtt_mode":"per_hop_half","delay_ms":100,"rate_mbps":20,"ifb_active":true}`)
	if err := CheckBenchTopology("ingress"); err != nil {
		t.Fatalf("accord ingress doit passer : %v", err)
	}
	writeTopo(t, p, `{"shape_mode":"egress","rtt_mode":"per_hop_half","delay_ms":100,"rate_mbps":20,"ifb_active":false}`)
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
	writeTopo(t, p, `{"shape_mode":"ingress","rtt_mode":"per_hop_half","delay_ms":100,"rate_mbps":20,"ifb_active":true}`)
	err := CheckBenchTopology("egress")
	if err == nil {
		t.Fatal("banc ingress + campagne egress doit être refusé")
	}
	if !strings.Contains(err.Error(), "ingress") || !strings.Contains(err.Error(), "egress") {
		t.Fatalf("l'erreur doit nommer les DEUX modes : %v", err)
	}
	writeTopo(t, p, `{"shape_mode":"egress","rtt_mode":"per_hop_half","delay_ms":100,"rate_mbps":20,"ifb_active":false}`)
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

func TestCheckBenchTopologyContractFieldsRefused(t *testing.T) {
	p := topoEnv(t)
	cases := []struct {
		name, topo, expected string
	}{
		{"rtt_mode invalide", `{"shape_mode":"egress","rtt_mode":"scalar","delay_ms":100,"rate_mbps":20,"ifb_active":false}`, "egress"},
		{"rtt_mode absent", `{"shape_mode":"egress","delay_ms":100,"rate_mbps":20,"ifb_active":false}`, "egress"},
		{"delay_ms nul", `{"shape_mode":"egress","rtt_mode":"per_hop_half","delay_ms":0,"rate_mbps":20,"ifb_active":false}`, "egress"},
		{"rate_mbps absent", `{"shape_mode":"egress","rtt_mode":"per_hop_half","delay_ms":100,"ifb_active":false}`, "egress"},
		{"ifb_active contredit ingress", `{"shape_mode":"ingress","rtt_mode":"per_hop_half","delay_ms":100,"rate_mbps":20,"ifb_active":false}`, "ingress"},
		{"ifb_active contredit egress", `{"shape_mode":"egress","rtt_mode":"per_hop_half","delay_ms":100,"rate_mbps":20,"ifb_active":true}`, "egress"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			writeTopo(t, p, tc.topo)
			if err := CheckBenchTopology(tc.expected); err == nil {
				t.Fatal("contrat invalide doit être refusé (échec fermé)")
			}
		})
	}
}
