package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Fabriques de TEST : jamais la preuve scellée du banc. Le producteur ne
// fabrique aucun compteur — il ne fait que recopier des entrées réelles déjà
// scellées. Ces chaînes ne vivent que dans le test.
func sealAuditJSON(source string, available bool, payload, drops, flows string) string {
	return fmt.Sprintf(`{"audit_id":"audit-20261122-abcdef12","kernel_telemetry":{"available":%t,"source":%q,"object_sha256":"object-sha-abc123","retrans_payload_bytes":%s,"qdisc_drop_packets":%s,"flows":%s}}`,
		available, source, payload, drops, flows)
}

const sealFlowsTwo = `[{"retrans_events":1},{"retrans_events":225}]`

func sealTcCapture(dropped uint64) string {
	return fmt.Sprintf("qdisc cake 1: dev e-wan root refcnt 2\n Sent 100 bytes 2 pkt (dropped %d, overlimits 0 requeues 0) \n backlog 0b 0p requeues 0\n", dropped)
}

func sealPutFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func sealReadDoc(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("sceau non écrit: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("sceau illisible: %v", err)
	}
	return doc
}

func sealNum(t *testing.T, doc map[string]any, key string, want float64) {
	t.Helper()
	got, ok := doc[key].(float64)
	if !ok || got != want {
		t.Fatalf("%s=%#v, attendu %v", key, doc[key], want)
	}
}

func sealStr(t *testing.T, doc map[string]any, key, want string) {
	t.Helper()
	if got, _ := doc[key].(string); got != want {
		t.Fatalf("%s=%#v, attendu %q", key, doc[key], want)
	}
}

// Oracle 1 — le sceau recopie EXACTEMENT les entrées réelles : aucune valeur
// fabriquée. 1+225 retrans_events, dropped 1 → 227, source ebpf, zéro qdisc.
// Le produit doit être consommable par loadSeal et donner match=true.
func TestBuildSealCopiesRealInputs(t *testing.T) {
	dir := t.TempDir()
	audit := sealPutFile(t, dir, "audit.json", sealAuditJSON("ebpf", true, "4096", "0", sealFlowsTwo))
	before := sealPutFile(t, dir, "before.txt", sealTcCapture(1))
	after := sealPutFile(t, dir, "after.txt", sealTcCapture(227))
	out := filepath.Join(dir, "seal.json")

	if err := buildSeal(audit, before, after, "e-wan", "cake", out); err != nil {
		t.Fatalf("buildSeal: %v", err)
	}
	doc := sealReadDoc(t, out)
	sealStr(t, doc, "schema", confrontSealSchema) // schéma exact, pas un dérivé
	sealStr(t, doc, "audit_id", "audit-20261122-abcdef12")
	sealStr(t, doc, "source", "ebpf")
	sealStr(t, doc, "object_sha256", "object-sha-abc123")
	sealNum(t, doc, "retrans_payload_bytes", 4096)
	sealNum(t, doc, "retrans_events", 226)
	sealNum(t, doc, "qdisc_drop_packets", 0)
	sealStr(t, doc, "tc_interface", "e-wan")
	sealStr(t, doc, "tc_qdisc", "cake")
	sealNum(t, doc, "tc_drops_before", 1)
	sealNum(t, doc, "tc_drops_after", 227)
	note, _ := doc["note"].(string)
	if !strings.Contains(note, "blind spot") || !strings.Contains(note, "not zero loss") {
		t.Fatalf("note absente sur le zéro qdisc: %q", note)
	}

	// Le produit est valide pour le consommateur : round-trip et confrontation.
	s, err := loadSeal(out)
	if err != nil {
		t.Fatalf("sceau produit refusé par loadSeal: %v", err)
	}
	if s.RetransEvents != 226 || s.TCDropsAfter-s.TCDropsBefore != s.RetransEvents {
		t.Fatalf("confrontation incohérente: %+v", s)
	}
	report := filepath.Join(dir, "report.json")
	if err := writeConfrontationReport(out, report); err != nil {
		t.Fatalf("writeConfrontationReport: %v", err)
	}
	rdoc := sealReadDoc(t, report)
	verdict, ok := rdoc["confrontation"].(map[string]any)
	if !ok || verdict["match"] != true {
		t.Fatalf("match attendu vrai: %#v", rdoc["confrontation"])
	}
}

// Oracle absence — l'audit réel 6.8-cake est sérialisé SANS qdisc_drop_packets
// (ancien omitempty). L'absence doit être représentée explicitement : drapeau
// vrai + note nommée, et surtout AUCUN compteur numérique fabriqué à zéro.
func TestBuildSealHonestAbsence(t *testing.T) {
	const auditNoQdisc = `{"audit_id":"audit-20261122-abcdef12","kernel_telemetry":{"available":true,"source":"ebpf","object_sha256":"object-sha-abc123","retrans_payload_bytes":4096,"flows":[{"retrans_events":1},{"retrans_events":225}]}}`
	dir := t.TempDir()
	aPath := sealPutFile(t, dir, "audit.json", auditNoQdisc)
	bPath := sealPutFile(t, dir, "before.txt", sealTcCapture(1))
	a2Path := sealPutFile(t, dir, "after.txt", sealTcCapture(227))
	out := filepath.Join(dir, "seal.json")

	if err := buildSeal(aPath, bPath, a2Path, "e-wan", "cake", out); err != nil {
		t.Fatalf("buildSeal devrait accepter l'absence explicite: %v", err)
	}
	doc := sealReadDoc(t, out)
	if doc["qdisc_drop_packets_absent"] != true {
		t.Fatalf("drapeau d'absence explicite manquant: %#v", doc["qdisc_drop_packets_absent"])
	}
	if v, ok := doc["qdisc_drop_packets"]; ok {
		t.Fatalf("compteur fabriqué malgré l'absence: qdisc_drop_packets=%#v", v)
	}
	note, _ := doc["note"].(string)
	if !strings.Contains(note, "blind spot") || !strings.Contains(note, "not a measured zero") {
		t.Fatalf("note d'absence nommée manquante: %q", note)
	}

	// Round-trip : le consommateur accepte ce sceau et prononce match=true
	// sans jamais réintroduire de compteur numérique.
	if _, err := loadSeal(out); err != nil {
		t.Fatalf("sceau d'absence refusé par loadSeal: %v", err)
	}
	report := filepath.Join(dir, "report.json")
	if err := writeConfrontationReport(out, report); err != nil {
		t.Fatalf("writeConfrontationReport (absence): %v", err)
	}
	rdoc := sealReadDoc(t, report)
	k, _ := rdoc["kernel_truth"].(map[string]any)
	if k["qdisc_drop_packets_absent"] != true {
		t.Fatalf("drapeau d'absence non recopié dans kernel_truth: %#v", k)
	}
	if _, ok := k["qdisc_drop_packets"]; ok {
		t.Fatalf("compteur fabriqué dans le rapport: %#v", k["qdisc_drop_packets"])
	}
	if knote, _ := k["note"].(string); knote != note {
		t.Fatalf("note non recopiée verbatim: %q", knote)
	}
	verdict, _ := rdoc["confrontation"].(map[string]any)
	if verdict["match"] != true {
		t.Fatalf("match attendu vrai (226 retrans == 226 delta tc): %#v", verdict)
	}
}

// Oracle hostile dédié — null est une absence déguisée, jamais une absence
// explicite : le producteur doit refuser fermé et ne rien publier.
func TestBuildSealRejectsNullQdiscDropPackets(t *testing.T) {
	dir := t.TempDir()
	aPath := sealPutFile(t, dir, "audit.json", sealAuditJSON("ebpf", true, "4096", "null", sealFlowsTwo))
	bPath := sealPutFile(t, dir, "before.txt", sealTcCapture(1))
	a2Path := sealPutFile(t, dir, "after.txt", sealTcCapture(227))
	out := filepath.Join(dir, "seal.json")
	err := buildSeal(aPath, bPath, a2Path, "e-wan", "cake", out)
	if err == nil || !strings.Contains(err.Error(), "qdisc_drop_packets") {
		t.Fatalf("qdisc_drop_packets=null doit être refusé: %v", err)
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Fatalf("fichier écrit malgré null: stat=%v", statErr)
	}
}

// Oracle 2 — chaque validation a un oracle hostile : supprimer la garde fait
// échouer ce test. Échec fermé (erreur + aucun fichier écrit) dans tous les cas.
func TestBuildSealFailsClosed(t *testing.T) {
	ok := sealAuditJSON("ebpf", true, "4096", "0", sealFlowsTwo)
	cases := []struct {
		name, auditStr, beforeStr, afterStr, wantReason string
		noAudit                                         bool
	}{
		{"audit absent", "", sealTcCapture(1), sealTcCapture(227), "audit", true},
		{"source tcpinfo", sealAuditJSON("tcpinfo", true, "4096", "0", sealFlowsTwo), sealTcCapture(1), sealTcCapture(227), "ebpf", false},
		{"télémétrie indisponible", sealAuditJSON("ebpf", false, "4096", "0", sealFlowsTwo), sealTcCapture(1), sealTcCapture(227), "available=false", false},
		{"aucun flux", sealAuditJSON("ebpf", true, "4096", "0", `[]`), sealTcCapture(1), sealTcCapture(227), "flows", false},
		{"compteur null", sealAuditJSON("ebpf", true, "null", "0", sealFlowsTwo), sealTcCapture(1), sealTcCapture(227), "retrans_payload_bytes", false},
		{"compteur qdisc null", sealAuditJSON("ebpf", true, "4096", "null", sealFlowsTwo), sealTcCapture(1), sealTcCapture(227), "qdisc_drop_packets", false},
		{"flux retrans_events null", sealAuditJSON("ebpf", true, "4096", "0", `[{"retrans_events":null}]`), sealTcCapture(1), sealTcCapture(227), "retrans_events", false},
		{"dropped dupliqué", ok, sealTcCapture(1) + sealTcCapture(2), sealTcCapture(227), "dropped", false},
		{"dropped absent", ok, "qdisc cake 1: dev e-wan root\n Sent 1 bytes 1 pkt (overlimits 0)\n", sealTcCapture(227), "dropped", false},
		{"after inférieur à before", ok, sealTcCapture(227), sealTcCapture(1), "drops_after", false},
		{"kernel_telemetry absent", `{"audit_id":"audit-x"}`, sealTcCapture(1), sealTcCapture(227), "kernel_telemetry", false},
		{"object_sha256 absent", `{"audit_id":"audit-x","kernel_telemetry":{"available":true,"source":"ebpf","retrans_payload_bytes":4096,"qdisc_drop_packets":0,"flows":[{"retrans_events":1}]}}`, sealTcCapture(1), sealTcCapture(227), "object_sha256", false},
		{"audit_id vide", `{"audit_id":"","kernel_telemetry":{"available":true,"source":"ebpf","object_sha256":"x","retrans_payload_bytes":4096,"qdisc_drop_packets":0,"flows":[{"retrans_events":1}]}}`, sealTcCapture(1), sealTcCapture(227), "audit_id", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			aPath := filepath.Join(dir, "audit.json")
			if !tc.noAudit {
				aPath = sealPutFile(t, dir, "audit.json", tc.auditStr)
			}
			bPath := sealPutFile(t, dir, "before.txt", tc.beforeStr)
			a2Path := sealPutFile(t, dir, "after.txt", tc.afterStr)
			out := filepath.Join(dir, "seal.json")

			err := buildSeal(aPath, bPath, a2Path, "e-wan", "cake", out)
			if err == nil {
				t.Fatal("échec fermé attendu, obtenu nil — validation supprimée ?")
			}
			if !strings.Contains(err.Error(), tc.wantReason) {
				t.Fatalf("raison %q absente de l'erreur: %v", tc.wantReason, err)
			}
			if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
				t.Fatalf("fichier écrit malgré le refus: stat=%v", statErr)
			}
		})
	}
}

// Oracle hostile — aucun chemin n'a de défaut : un chemin vide est refusé,
// pas remplacé par une entrée implicite.
func TestBuildSealRejectsEmptyPaths(t *testing.T) {
	dir := t.TempDir()
	f := sealPutFile(t, dir, "x.txt", "x")
	out := filepath.Join(dir, "seal.json")
	if err := buildSeal("", f, f, "e-wan", "cake", out); err == nil || !strings.Contains(err.Error(), "requis") {
		t.Fatalf("chemin vide accepté: %v", err)
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Fatalf("fichier écrit depuis un chemin vide: stat=%v", statErr)
	}
}

// Oracle hostile — le schéma du sceau est la racine de confiance du
// consommateur : un schéma erroné doit être refusé (et le producteur l'écrit
// toujours exactement, voir Oracle 1).
func TestLoadSealRejectsWrongSchema(t *testing.T) {
	dir := t.TempDir()
	bad := `{"schema":"cgo.kernel-truth-seal/2","audit_id":"x","source":"ebpf","retrans_events":1,"qdisc_drop_packets":0,"tc_drops_before":0,"tc_drops_after":1}`
	p := sealPutFile(t, dir, "bad.json", bad)
	if _, err := loadSeal(p); err == nil || !strings.Contains(err.Error(), "schema") {
		t.Fatalf("schéma erroné accepté ou mauvaise raison: %v", err)
	}
}

// Oracle CLI — une commande sans audit sort 2 et n'écrit aucun fichier.
func TestKernelTruthSealCmdMissingAuditExits2(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "seal.json")
	code := runKernelTruthSealCmd([]string{
		"--audit", filepath.Join(dir, "absent.json"),
		"--tc-before", sealPutFile(t, dir, "before.txt", sealTcCapture(1)),
		"--tc-after", sealPutFile(t, dir, "after.txt", sealTcCapture(227)),
		"--iface", "e-wan", "--qdisc", "cake", "--out", out,
	})
	if code != 2 {
		t.Fatalf("code=%d, attendu 2", code)
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Fatalf("fichier écrit malgré l'absence d'audit: stat=%v", statErr)
	}
}

// Oracle — un capture `tc -s qdisc show` réel contient PLUSIEURS qdisc (ici 4),
// chacun avec une ligne `dropped`. L'ambiguïté se juge dans le bloc cible, pas
// sur tout le fichier : la garde globale dupliquée refusait ce capture à tort.
func TestParseTCDropsMultiInterfaceCapture(t *testing.T) {
	const multi = "qdisc cake 1: dev w-edge root refcnt 2\n Sent 100 bytes 2 pkt (dropped 3, overlimits 0 requeues 0)\n backlog 0b 0p requeues 0\n" +
		"qdisc cake 1: dev c-wan root refcnt 2\n Sent 200 bytes 4 pkt (dropped 7, overlimits 0 requeues 0)\n backlog 0b 0p requeues 0\n" +
		"qdisc cake 1: dev e-wan root refcnt 2\n Sent 300 bytes 6 pkt (dropped 226, overlimits 0 requeues 0)\n backlog 0b 0p requeues 0\n" +
		"qdisc fq_codel 2: dev e-srv root refcnt 2\n Sent 400 bytes 8 pkt (dropped 11, overlimits 0 requeues 0)\n backlog 0b 0p requeues 0\n"
	p := sealPutFile(t, t.TempDir(), "multi.txt", multi)

	got, err := parseTCDrops(p, "e-wan", "cake")
	if err != nil {
		t.Fatalf("capture multi-interface valide refusée: %v", err)
	}
	if got != 226 {
		t.Fatalf("dropped=%d, attendu 226 (bloc cake dev e-wan)", got)
	}

	// Hostile : mauvais iface ou mauvais qdisc → échec fermé, même si une
	// valeur `dropped` existe ailleurs dans le capture.
	for _, tc := range []struct{ iface, qdisc string }{
		{"w-edge2", "cake"},
		{"e-wan", "fq_codel"},
	} {
		if _, err := parseTCDrops(p, tc.iface, tc.qdisc); err == nil {
			t.Fatalf("labels %s/%s acceptés à tort", tc.iface, tc.qdisc)
		}
	}
}

// Oracle hostile — même avec plusieurs blocs, deux `dropped` DANS le bloc
// cible restent ambigus : le contrôle par bloc n'est pas relâché.
func TestParseTCDropsMultiInterfaceAmbiguousTarget(t *testing.T) {
	const multi = "qdisc cake 1: dev e-wan root refcnt 2\n Sent 1 pkt (dropped 1, overlimits 0)\n Sent 2 pkt (dropped 2, overlimits 0)\n" +
		"qdisc cake 1: dev c-wan root refcnt 2\n Sent 3 pkt (dropped 3, overlimits 0)\n"
	p := sealPutFile(t, t.TempDir(), "multi-ambiguous.txt", multi)
	if _, err := parseTCDrops(p, "e-wan", "cake"); err == nil {
		t.Fatal("deux dropped dans le bloc cible acceptés — garde par bloc supprimée ?")
	}
}

// Oracle hostile — --iface/--qdisc sont des étiquettes demandées, pas des
// faits. Le capture doit les confirmer lui-même : un capture d'un autre iface
// ou d'un autre qdisc, même avec un seul `dropped`, ne doit jamais être
// recollé sous les labels demandés (fail fermé, aucun fichier publié).
func TestBuildSealVerifiesTCLabels(t *testing.T) {
	ok := sealAuditJSON("ebpf", true, "4096", "0", sealFlowsTwo)
	const (
		iface = "e-wan"
		qdisc = "cake"
	)
	cases := []struct {
		name, beforeStr, afterStr, wantReason string
		wantOK                                bool
	}{
		{
			name:      "labels exacts (positif)",
			beforeStr: sealTcCapture(1),
			afterStr:  sealTcCapture(227),
			wantOK:    true,
		},
		{
			name:       "iface erroné",
			beforeStr:  "qdisc cake 1: dev e-wan2 root refcnt 2\n Sent 1 pkt (dropped 1, overlimits 0)\n",
			afterStr:   "qdisc cake 1: dev e-wan2 root refcnt 2\n Sent 1 pkt (dropped 227, overlimits 0)\n",
			wantReason: "iface",
		},
		{
			name:       "qdisc erroné",
			beforeStr:  "qdisc fq_codel 1: dev e-wan root refcnt 2\n Sent 1 pkt (dropped 1, overlimits 0)\n",
			afterStr:   "qdisc fq_codel 1: dev e-wan root refcnt 2\n Sent 1 pkt (dropped 227, overlimits 0)\n",
			wantReason: "qdisc",
		},
		{
			name:       "ligne dev absente",
			beforeStr:  "qdisc cake 1: root refcnt 2\n Sent 1 pkt (dropped 1, overlimits 0)\n",
			afterStr:   "qdisc cake 1: root refcnt 2\n Sent 1 pkt (dropped 227, overlimits 0)\n",
			wantReason: "dev",
		},
		{
			name: "dropped sur un autre bloc que la cible",
			beforeStr: "qdisc fq_codel 1: dev e-wan root refcnt 2\n Sent 1 pkt (dropped 1, overlimits 0)\n" +
				"qdisc cake 2: dev e-wan root refcnt 2\n Sent 1 pkt (overlimits 0)\n",
			afterStr:   sealTcCapture(227),
			wantReason: "dropped",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			aPath := sealPutFile(t, dir, "audit.json", ok)
			bPath := sealPutFile(t, dir, "before.txt", tc.beforeStr)
			a2Path := sealPutFile(t, dir, "after.txt", tc.afterStr)
			out := filepath.Join(dir, "seal.json")

			err := buildSeal(aPath, bPath, a2Path, iface, qdisc, out)
			if tc.wantOK {
				if err != nil {
					t.Fatalf("labels exacts refusés à tort: %v", err)
				}
				doc := sealReadDoc(t, out)
				sealNum(t, doc, "tc_drops_before", 1)
				sealNum(t, doc, "tc_drops_after", 227)
				return
			}
			if err == nil {
				t.Fatal("capture aux labels divergents acceptée — vérification supprimée ?")
			}
			if !strings.Contains(err.Error(), tc.wantReason) {
				t.Fatalf("raison %q absente de l'erreur: %v", tc.wantReason, err)
			}
			if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
				t.Fatalf("fichier écrit malgré le refus: stat=%v", statErr)
			}
		})
	}
}
