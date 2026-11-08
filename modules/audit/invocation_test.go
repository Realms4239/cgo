package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Realms4239/cgo/remediate"
)

// La promesse de la console : une donnée normalisée mais bornée par la donnée
// brute et la commande exacte. Le rapport scellait des MESURES et la provenance
// du CSV, jamais les réglages qui avaient produit la mesure — le tiroir n'avait
// donc rien à montrer comme commande, et un vérificateur ne pouvait pas relancer
// le même audit. Ces tests gèlent l'invocation dans le moteur, à l'endroit où
// les défauts sont déjà résolus, et non dans la surface qui les a reçus.

// TestRunSealsTheEffectiveInvocation — le moteur scelle ce qu'il a UTILISÉ :
// une direction vide devient "up", des flux à zéro deviennent 4. Recopier la
// demande telle quelle écrirait une commande qui ne reproduit pas la mesure.
func TestRunSealsTheEffectiveInvocation(t *testing.T) {
	d := Deps{
		Ping:  func(context.Context, string, int) []float64 { return []float64{20, 21, 22, 23, 30} },
		Small: func(context.Context) (float64, error) { return 0, fmt.Errorf("no small URL") },
		BulkLoad: func(_ context.Context, _, dir string, flows int) LoadOutcome {
			return LoadOutcome{Bytes: 1_000_000, PerFlow: make([]uint64, flows), Elapsed: time.Second}
		},
	}
	res, err := Run(context.Background(), Params{
		AuditID:  "audit-invocation",
		Duration: 2,
		Site:     "Départ 22",
		LinkType: "4g",
		Target:   "10.30.0.2",
		SmallURL: "http://10.30.0.2:8081/small",
		BulkAddr: "10.30.0.2:5201",
		SinkMode: "dualns",
		Origin:   "cli",
	}, d)
	if err != nil {
		t.Fatal(err)
	}
	in := res.Invocation
	if in.Direction != "up" {
		t.Errorf("direction vide normalisée puis scellée : %q", in.Direction)
	}
	if in.Flows != 4 {
		t.Errorf("flux à zéro = le défaut du moteur (4), scellé comme tel : %d", in.Flows)
	}
	if in.Duration != 2 || in.SinkMode != "dualns" || in.Origin != "cli" {
		t.Errorf("réglages effectifs : %+v", in)
	}
	// Chaque réglage qui change la mesure est explicite : une reproduction qui
	// laisse -flows hors de la ligne re-détecterait le banc au lieu de le rejouer.
	for _, want := range []string{
		"cgo audit", "-duration=2", "-flows=4", "-direction='up'",
		"-site='Départ 22'", "-link-type='4g'",
		"-target='10.30.0.2'", "-small='http://10.30.0.2:8081/small'",
		"-bulk='10.30.0.2:5201'",
	} {
		if !strings.Contains(in.Command, want) {
			t.Errorf("commande de reproduction sans %s : %s", want, in.Command)
		}
	}
}

// TestInvocationQuotesEveryValue — la ligne est recopiée par une personne dans
// un shell POSIX. Une apostrophe dans un nom de site doit sortir échappée, sinon
// le reste de la chaîne devient une commande que le moteur n'a jamais voulue.
func TestInvocationQuotesEveryValue(t *testing.T) {
	in := invocationFrom(Params{
		Duration: 60, Flows: 4, Direction: "up", Target: "8.8.8.8",
		Site: `A'; rm -rf /; echo '`,
	})
	want := `-site='A'\''; rm -rf /; echo '\''`
	if !strings.Contains(in.Command, want) {
		t.Fatalf("échappement POSIX de l'apostrophe attendu (%s) dans : %s", want, in.Command)
	}
	// Une valeur vide ne devient pas un argument vide : -bulk='' demanderait
	// explicitement « pas de charge », ce qui n'est pas ce qui a tourné.
	if strings.Contains(in.Command, "-bulk") || strings.Contains(in.Command, "-small") {
		t.Fatalf("réglage absent publié comme vide : %s", in.Command)
	}
}

// TestBuildReportSealsReproduction — le rapport porte la commande, y compris les
// deux réglages qui ne vivent qu'ici (qdisc choisi, asymmetric, cpu-constrained),
// et le JSON expose le champ sous un nom stable pour la console.
func TestBuildReportSealsReproduction(t *testing.T) {
	res := &Result{
		AuditID:   "audit-repro",
		Timestamp: "2026-11-01T10:00:00Z",
		Site:      "tana", LinkType: "4g", Provider: "yas",
		RTTIdleP95: 49, RTTLoadedP95: 240, LoadAttempted: true,
		Invocation: invocationFrom(Params{
			Duration: 120, Flows: 4, Direction: "both", Site: "tana",
			LinkType: "4g", Provider: "yas", Target: "10.200.0.1",
			SinkMode: "legacy", Origin: "http",
		}),
	}
	rep, err := BuildReport(res, remediate.Link{Interface: "eth0", RouterOSInterface: "ether1-wan"}, "cake", false, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"-qdisc='cake'", "-asymmetric=false", "-cpu-constrained", "-direction='both'", "-iface='eth0'", "-routeros-iface='ether1-wan'"} {
		if !strings.Contains(rep.Invocation.Command, want) {
			t.Errorf("commande sans %s : %s", want, rep.Invocation.Command)
		}
	}
	// -sink-mode n'existe pas comme drapeau : le mode de détection est un FAIT
	// consigné pour le tiroir, jamais un argument recollable dans un shell.
	if strings.Contains(rep.Invocation.Command, "-sink-mode") {
		t.Errorf("la commande publie un drapeau que le CLI refuse : %s", rep.Invocation.Command)
	}
	if rep.Invocation.SinkMode != "legacy" || rep.Invocation.Origin != "http" {
		t.Errorf("détection et origine consignées : %+v", rep.Invocation)
	}
	// Un audit jamais chiffré garde sa commande : la reproduction ne dépend pas
	// d'une mesure réussie, sinon le tiroir n'aurait rien à montrer sur un échec.
	b, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"invocation"`) || !strings.Contains(string(b), `"replay_command"`) {
		t.Errorf("contrat JSON du rapport : invocation/replay_command absents")
	}
}

// TestArchivedReportRendersAnAbsence — un rapport scellé avant ce champ se relit
// avec une invocation vide, et le Markdown doit le dire au lieu d'imprimer une
// commande inventée. C'est la même règle que pour le verdict opérateur.
func TestArchivedReportRendersAnAbsence(t *testing.T) {
	res := &Result{
		AuditID: "audit-vieux", Timestamp: "2026-10-30T10:00:00Z",
		Site: "tana", LinkType: "4g", RTTIdleP95: 49, RTTLoadedP95: 240,
		LoadAttempted: true,
	}
	rep, err := BuildReport(res, remediate.Link{Interface: "eth0"}, "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Invocation != nil {
		t.Fatalf("un résultat sans invocation ne doit pas fabriquer le bloc : %+v", rep.Invocation)
	}
	md := rep.RenderMarkdown()
	if !strings.Contains(md, "Reproduction") || !strings.Contains(md, "non enregistr") {
		t.Errorf("markdown d'un archive sans invocation : %s", md)
	}
	if strings.Contains(md, "cgo audit") {
		t.Errorf("le markdown invente une reproduction : %s", md)
	}
	// Le bloc absent ne se publie pas vide : un « flows: 0 » lu comme « un flux
	// demandé » est exactement le faux vert que le projet refuse depuis 20A.
	b, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"invocation"`) {
		t.Fatalf("invocation publiée comme un bloc zéré au lieu d'une absence : %s", b)
	}
}
