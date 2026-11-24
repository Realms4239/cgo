// Oracles du contrat de topologie : le check ne compare qu'un contrat déclaré
// à des faits fournis. Il n'invente aucun fait, n'exécute aucun ip/tc et ne
// mutation rien. Un champ absent est un drift explicite, jamais un accord.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Realms4239/cgo/pkg/campagne"
)

// validContract reflète la topologie réellement bâtie par kit/run-dualns.sh
// (4 namespaces, 6 interfaces de transit, cake, 20 Mbps, 100 ms RTT, egress).
const validContract = `{
  "namespaces": ["ns_client", "ns_wan_emulator", "ns_edge_router", "ns_server"],
  "ifaces": ["c-cli", "c-wan", "w-edge", "e-wan", "e-srv", "s-edge"],
  "qdisc": "cake",
  "cap_mbps": 20,
  "delay_ms": 100,
  "shape_mode": "egress"
}`

func writeTopoFile(t *testing.T, name, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func mutatedContract(t *testing.T, mutate func(map[string]any)) string {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(validContract), &m); err != nil {
		t.Fatal(err)
	}
	mutate(m)
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func matchingFacts() Facts {
	return Facts{
		Namespaces: []string{"ns_client", "ns_wan_emulator", "ns_edge_router", "ns_server"},
		Ifaces:     []string{"c-cli", "c-wan", "w-edge", "e-wan", "e-srv", "s-edge"},
		Qdisc:      "cake",
		CapMbps:    20,
		DelayMs:    100,
		ShapeMode:  "egress",
	}
}

func driftFor(drifts []Drift, field string) (Drift, bool) {
	for _, d := range drifts {
		if d.Field == field {
			return d, true
		}
	}
	return Drift{}, false
}

// TestCheckShapeModeDrift — oracle hostile : un contrat ingress face à des
// faits egress doit produire un drift explicite (fausse certification refusée).
func TestCheckShapeModeDrift(t *testing.T) {
	c, err := Load(writeTopoFile(t, "contract.json", mutatedContract(t, func(m map[string]any) {
		m["shape_mode"] = "ingress"
	})))
	if err != nil {
		t.Fatal(err)
	}
	drifts := Check(c, matchingFacts())
	d, ok := driftFor(drifts, "shape_mode")
	if !ok {
		t.Fatalf("contrat ingress / faits egress : aucun drift shape_mode : %v", drifts)
	}
	if d.Declared != "ingress" || d.Observed != "egress" {
		t.Fatalf("drift shape_mode = %+v, want ingress/egress", d)
	}
}

// TestCheckAgreementNoDrift — oracle positif : des faits conformes ne
// produisent aucun drift (le check ne signale pas de faux positifs).
func TestCheckAgreementNoDrift(t *testing.T) {
	c, err := Load(writeTopoFile(t, "contract.json", validContract))
	if err != nil {
		t.Fatal(err)
	}
	if drifts := Check(c, matchingFacts()); len(drifts) != 0 {
		t.Fatalf("faits conformes : %d drifts, want 0 : %v", len(drifts), drifts)
	}
}

// TestLoadMissingFileRefuses — un contrat absent est un échec fermé, pas un
// accord tacite.
func TestLoadMissingFileRefuses(t *testing.T) {
	p := filepath.Join(t.TempDir(), "absent.json")
	if _, err := Load(p); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("contrat absent : err = %v, want os.ErrNotExist", err)
	}
}

// TestLoadStrictRefuses — décodage strict : clé inconnue, champ manquant,
// liste vide, shape_mode invalide et valeurs JSON multiples sont refusés.
func TestLoadStrictRefuses(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"unknown field", mutatedContract(t, func(m map[string]any) { m["surprise"] = 1 })},
		{"missing qdisc", mutatedContract(t, func(m map[string]any) { delete(m, "qdisc") })},
		{"empty namespaces", mutatedContract(t, func(m map[string]any) { m["namespaces"] = []string{} })},
		{"bad shape mode", mutatedContract(t, func(m map[string]any) { m["shape_mode"] = "sideways" })},
		{"zero capacity", mutatedContract(t, func(m map[string]any) { m["cap_mbps"] = 0 })},
		{"trailing value", validContract + "\n{}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Load(writeTopoFile(t, "contract.json", tc.body)); err == nil {
				t.Fatal("contrat invalide accepté")
			}
		})
	}
}

// TestCheckAbsentFactsExplicit — un fait absent est signalé comme tel, jamais
// supposé égal au contrat.
func TestCheckAbsentFactsExplicit(t *testing.T) {
	c, err := Load(writeTopoFile(t, "contract.json", validContract))
	if err != nil {
		t.Fatal(err)
	}
	facts := Facts{Qdisc: "cake"}
	drifts := Check(c, facts)
	for _, field := range []string{"namespaces", "ifaces", "cap_mbps", "delay_ms", "shape_mode"} {
		d, ok := driftFor(drifts, field)
		if !ok || d.Observed != "(absent)" {
			t.Fatalf("fait %s absent : drift = %+v (présent=%v), want observed '(absent)'", field, d, ok)
		}
	}
}

// TestFingerprintStableAndSensitive — empreinte stable entre deux lectures et
// sensible à toute altération du contrat (oracle hostile).
func TestFingerprintStableAndSensitive(t *testing.T) {
	p := writeTopoFile(t, "contract.json", validContract)
	first, err := Fingerprint(p)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Fingerprint(p)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("empreinte instable : %s puis %s", first, second)
	}
	sum := sha256.Sum256([]byte(validContract))
	want := hex.EncodeToString(sum[:])
	if first != want {
		t.Fatalf("empreinte = %s, want SHA-256 des octets du contrat %s", first, want)
	}
	tampered := writeTopoFile(t, "tampered.json", mutatedContract(t, func(m map[string]any) {
		m["cap_mbps"] = 200
	}))
	got, err := Fingerprint(tampered)
	if err != nil {
		t.Fatal(err)
	}
	if got == first {
		t.Fatal("contrat altéré : empreinte inchangée")
	}
}

// TestFreezeManifestCarriesContractFingerprint — le manifeste porte un champ
// ADDITIF dont la valeur est l'empreinte du contrat déclaré ; aucun champ
// existant n'est modifié.
func TestFreezeManifestCarriesContractFingerprint(t *testing.T) {
	contractPath := writeTopoFile(t, "topology.contract.json", validContract)
	t.Setenv("CGO_TOPOLOGY_CONTRACT", contractPath)
	dir := t.TempDir()
	w, err := campagne.OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := []byte(`{"profiles":["P2"],"reps":2}`)
	if err := w.Freeze(cfg); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(validContract))
	want := hex.EncodeToString(sum[:])
	if got, _ := doc["topology_contract_sha256"].(string); got != want {
		t.Fatalf("topology_contract_sha256 = %q, want %q", got, want)
	}
	for _, key := range []string{"config_sha256", "config", "files"} {
		if _, ok := doc[key]; !ok {
			t.Fatalf("champ existant %s supprimé du manifeste", key)
		}
	}
}

// TestFreezeWithoutContractOmitsFingerprint — pas de contrat configurable ⇒
// aucun hash factice dans le manifeste.
func TestFreezeWithoutContractOmitsFingerprint(t *testing.T) {
	t.Setenv("CGO_TOPOLOGY_CONTRACT", "")
	dir := t.TempDir()
	w, err := campagne.OpenRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Freeze([]byte(`{"test":true}`)); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	if _, ok := doc["topology_contract_sha256"]; ok {
		t.Fatal("empreinte de contrat présente alors qu'aucun contrat n'était configuré")
	}
}

// TestFreezeUnreadableContractRefuses — un contrat explicitement configuré mais
// illisible fait échouer le gel (échec fermé).
func TestFreezeUnreadableContractRefuses(t *testing.T) {
	t.Setenv("CGO_TOPOLOGY_CONTRACT", filepath.Join(t.TempDir(), "absent.json"))
	w, err := campagne.OpenRun(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Freeze([]byte(`{}`)); err == nil {
		t.Fatal("contrat configuré mais illisible : gel accepté")
	}
}
