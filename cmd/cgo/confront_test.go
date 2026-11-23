package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureSeal est une entrée de TEST, jamais la preuve scellée du banc. Le
// binaire ne fabrique aucun compteur : il ne fait que recopier ce qu'il lit.
const fixtureSeal = `{
  "schema": "cgo.kernel-truth-seal/1",
  "audit_id": "audit-fixture-0001",
  "source": "ebpf",
  "object_sha256": "fixture-object-sha256",
  "retrans_payload_bytes": 1024,
  "retrans_events": 7,
  "qdisc_drop_packets": 0,
  "tc_interface": "e-wan",
  "tc_qdisc": "cake",
  "tc_drops_before": 2,
  "tc_drops_after": 9,
  "note": "TEST FIXTURE — not the sealed bench evidence"
}`

func putSeal(t *testing.T, dir, body string) string {
	t.Helper()
	p := filepath.Join(dir, "sealed-input.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func readReport(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("rapport non écrit: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("rapport illisible: %v", err)
	}
	return doc
}

func num(t *testing.T, m map[string]any, key string) float64 {
	t.Helper()
	v, ok := m[key].(float64)
	if !ok {
		t.Fatalf("%s absent ou non numérique: %#v", key, m[key])
	}
	return v
}

func obj(t *testing.T, m map[string]any, key string) map[string]any {
	t.Helper()
	v, ok := m[key].(map[string]any)
	if !ok {
		t.Fatalf("%s absent ou non objet: %#v", key, m[key])
	}
	return v
}

// Oracle 1 — le drapeau absent n'écrit jamais le fichier (attendu déterministe).
func TestConfrontOutFileAbsentWithoutFlag(t *testing.T) {
	dir := t.TempDir()
	in := putSeal(t, dir, fixtureSeal)
	out := filepath.Join(dir, "ebpf_vs_tc_report.json")
	if code := runConfrontOut("", in); code != 0 {
		t.Fatalf("drapeau absent: code=%d, attendu 0", code)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("fichier écrit sans --confront-out: stat=%v", err)
	}
}

// Oracle 2 — le rapport recopie exactement le sceau lu, calcule le delta tc,
// et prononce l'accord sans inventer de valeur.
func TestConfrontOutWritesSealedTruth(t *testing.T) {
	dir := t.TempDir()
	in := putSeal(t, dir, fixtureSeal)
	out := filepath.Join(dir, "ebpf_vs_tc_report.json")
	if code := runConfrontOut(out, in); code != 0 {
		t.Fatalf("code=%d, attendu 0", code)
	}
	doc := readReport(t, out)
	if doc["schema"] != "cgo.ebpf-vs-tc/1" {
		t.Fatalf("schema=%#v", doc["schema"])
	}
	if doc["audit_id"] != "audit-fixture-0001" || doc["source"] != "ebpf" {
		t.Fatalf("provenance non recopiée: %#v", doc)
	}
	k := obj(t, doc, "kernel_truth")
	if num(t, k, "retrans_events") != 7 || num(t, k, "qdisc_drop_packets") != 0 || num(t, k, "retrans_payload_bytes") != 1024 {
		t.Fatalf("compteurs noyau altérés: %#v", k)
	}
	if k["object_sha256"] != "fixture-object-sha256" {
		t.Fatalf("object_sha256 altéré: %#v", k["object_sha256"])
	}
	tc := obj(t, doc, "tc_s")
	if num(t, tc, "drops_before") != 2 || num(t, tc, "drops_after") != 9 || num(t, tc, "drops_delta") != 7 {
		t.Fatalf("compteurs tc altérés: %#v", tc)
	}
	c := obj(t, doc, "confrontation")
	if num(t, c, "ebpf_retrans_events") != 7 || num(t, c, "tc_drops_delta") != 7 || c["match"] != true {
		t.Fatalf("confrontation fausse: %#v", c)
	}
}

// Oracle absence — le consommateur accepte un sceau dont le compteur est
// explicitement absent (drapeau vrai, aucun champ numérique).
func TestLoadSealAcceptsExplicitAbsence(t *testing.T) {
	dir := t.TempDir()
	absent := `{"schema":"cgo.kernel-truth-seal/1","audit_id":"audit-absent","source":"ebpf","retrans_events":226,"qdisc_drop_packets_absent":true,"tc_drops_before":1,"tc_drops_after":227,"note":"absence explicite"}`
	p := putSeal(t, dir, absent)
	if _, err := loadSeal(p); err != nil {
		t.Fatalf("sceau d'absence explicite refusé: %v", err)
	}
}

// Oracle hostile — ni compteur numérique ni drapeau d'absence vrai : refus
// fermé. Un drapeau false ne compte PAS comme une absence explicite.
func TestLoadSealRejectsNeitherNumericNorAbsence(t *testing.T) {
	cases := []string{
		`{"schema":"cgo.kernel-truth-seal/1","audit_id":"x","source":"ebpf","retrans_events":226,"tc_drops_before":1,"tc_drops_after":227}`,
		`{"schema":"cgo.kernel-truth-seal/1","audit_id":"x","source":"ebpf","retrans_events":226,"qdisc_drop_packets_absent":false,"tc_drops_before":1,"tc_drops_after":227}`,
	}
	for _, body := range cases {
		dir := t.TempDir()
		p := putSeal(t, dir, body)
		if _, err := loadSeal(p); err == nil {
			t.Fatal("sceau sans compteur ni absence accepté")
		}
	}
}

// Oracle — le rapport nomme les deux plans de vérité sans hiérarchie : tc est
// la surface comptable, eBPF la surface temporelle. Champ additif, aucun
// compteur ni version de schéma n'est modifié.
func TestConfrontReportNamesTruthPlanesWithoutHierarchy(t *testing.T) {
	dir := t.TempDir()
	in := putSeal(t, dir, fixtureSeal)
	out := filepath.Join(dir, "ebpf_vs_tc_report.json")
	if code := runConfrontOut(out, in); code != 0 {
		t.Fatalf("code=%d, attendu 0", code)
	}
	doc := readReport(t, out)
	if doc["schema"] != "cgo.ebpf-vs-tc/1" {
		t.Fatalf("schema altéré: %#v", doc["schema"])
	}
	tp := obj(t, doc, "truth_planes")
	tcRole, ok := tp["tc_qdisc"].(string)
	if !ok || !strings.Contains(strings.ToLower(tcRole), "accounting") {
		t.Fatalf("tc/qdisc n'est pas nommé surface comptable: %#v", tp["tc_qdisc"])
	}
	ebpfRole, ok := tp["ebpf"].(string)
	if !ok || !strings.Contains(strings.ToLower(ebpfRole), "temporal") {
		t.Fatalf("eBPF n'est pas nommé surface temporelle: %#v", tp["ebpf"])
	}
	blob, err := json.Marshal(tp)
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(blob))
	for _, w := range []string{"primary", "hierarchy", "secondary", "demot", "subordinat"} {
		if strings.Contains(lower, w) {
			t.Fatalf("claim hiérarchique interdit %q dans truth_planes: %s", w, blob)
		}
	}
}

// Oracle absence — le rapport recopie drapeau + note verbatim, n'invente aucun
// compteur, et prononce match=true depuis retrans_events vs delta tc.
func TestConfrontOutHonestAbsence(t *testing.T) {
	const absentSeal = `{
  "schema": "cgo.kernel-truth-seal/1",
  "audit_id": "audit-fixture-absent",
  "source": "ebpf",
  "retrans_events": 226,
  "qdisc_drop_packets_absent": true,
  "tc_drops_before": 1,
  "tc_drops_after": 227,
  "note": "TEST FIXTURE — explicit absence, not a measured zero"
}`
	dir := t.TempDir()
	in := putSeal(t, dir, absentSeal)
	out := filepath.Join(dir, "ebpf_vs_tc_report.json")
	if code := runConfrontOut(out, in); code != 0 {
		t.Fatalf("code=%d, attendu 0", code)
	}
	doc := readReport(t, out)
	k := obj(t, doc, "kernel_truth")
	if k["qdisc_drop_packets_absent"] != true {
		t.Fatalf("drapeau d'absence non recopié: %#v", k)
	}
	if _, ok := k["qdisc_drop_packets"]; ok {
		t.Fatalf("compteur fabriqué dans le rapport: %#v", k["qdisc_drop_packets"])
	}
	if k["note"] != "TEST FIXTURE — explicit absence, not a measured zero" {
		t.Fatalf("note non recopiée verbatim: %#v", k["note"])
	}
	c := obj(t, doc, "confrontation")
	if num(t, c, "ebpf_retrans_events") != 226 || num(t, c, "tc_drops_delta") != 226 || c["match"] != true {
		t.Fatalf("match attendu vrai depuis retrans vs delta tc: %#v", c)
	}
}

// Oracle hostile — sceau absent : échec fermé (code 2), aucun fichier écrit.
func TestConfrontOutFailsClosedWhenSealAbsent(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "ebpf_vs_tc_report.json")
	in := filepath.Join(dir, "pas-de-sceau.json")
	if code := runConfrontOut(out, in); code != 2 {
		t.Fatalf("sceau absent: code=%d, attendu 2", code)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("fichier écrit malgré l'absence du sceau: stat=%v", err)
	}
}

// Oracle hostile — un sceau qui n'est pas ebpf est refusé, pas recopié.
func TestConfrontOutRejectsNonEbpfSeal(t *testing.T) {
	dir := t.TempDir()
	bad := `{"schema":"cgo.kernel-truth-seal/1","audit_id":"x","source":"tcpinfo","retrans_events":7,"qdisc_drop_packets":0,"tc_drops_before":2,"tc_drops_after":9}`
	in := putSeal(t, dir, bad)
	out := filepath.Join(dir, "ebpf_vs_tc_report.json")
	if code := runConfrontOut(out, in); code != 2 {
		t.Fatalf("sceau tcpinfo: code=%d, attendu 2", code)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("fichier écrit depuis un sceau non-ebpf: stat=%v", err)
	}
}

// Oracle hostile — un sceau dont les compteurs sont absents est refusé.
func TestConfrontOutRejectsMissingCounters(t *testing.T) {
	dir := t.TempDir()
	bad := `{"schema":"cgo.kernel-truth-seal/1","audit_id":"x","source":"ebpf"}`
	in := putSeal(t, dir, bad)
	out := filepath.Join(dir, "ebpf_vs_tc_report.json")
	if code := runConfrontOut(out, in); code != 2 {
		t.Fatalf("compteurs absents: code=%d, attendu 2", code)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("fichier écrit depuis un sceau incomplet: stat=%v", err)
	}
}

// Oracle hostile — un compteur numérique requis à null doit être refusé.
// Une garde de présence seule laisse json.Unmarshal transformer null en 0
// silencieusement : c'est le trou absence-never-zero. Échec fermé, aucun fichier.
func TestConfrontOutRejectsNullRequiredCounter(t *testing.T) {
	for _, key := range []string{"retrans_events", "qdisc_drop_packets", "tc_drops_before", "tc_drops_after"} {
		t.Run(key, func(t *testing.T) {
			dir := t.TempDir()
			// Clé dupliquée : la dernière valeur (null) fait foi pour l'encodeur.
			bad := fmt.Sprintf(`{"schema":"cgo.kernel-truth-seal/1","audit_id":"x","source":"ebpf","retrans_events":7,"qdisc_drop_packets":0,"tc_drops_before":2,"tc_drops_after":9,"%s":null}`, key)
			in := putSeal(t, dir, bad)
			out := filepath.Join(dir, "ebpf_vs_tc_report.json")
			if code := runConfrontOut(out, in); code != 2 {
				t.Fatalf("%s=null: code=%d, attendu 2", key, code)
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatalf("fichier écrit depuis un sceau avec %s=null: stat=%v", key, err)
			}
		})
	}
}
