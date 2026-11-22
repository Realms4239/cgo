package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Helpers locaux : ce fichier est autonome et reproductible — aucun helper
// défini ailleurs n'est requis pour le lire ou l'exécuter.

func qdiscDoc(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("rapport non écrit (%s): %v", path, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("rapport illisible (%s): %v", path, err)
	}
	return doc
}

func qdiscObj(t *testing.T, doc map[string]any, key string) map[string]any {
	t.Helper()
	v, ok := doc[key].(map[string]any)
	if !ok {
		t.Fatalf("%s absent ou non objet: %#v", key, doc[key])
	}
	return v
}

func qdiscNum(t *testing.T, doc map[string]any, key string) float64 {
	t.Helper()
	v, ok := doc[key].(float64)
	if !ok {
		t.Fatalf("%s absent ou non numérique: %#v", key, doc[key])
	}
	return v
}

func qdiscSeal(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// Deux fixtures EXPLICITES : aucun strings.Replace, donc aucun risque de
// désaccord d'espacement entre la valeur scellée et la valeur recherchée.
const qdiscSealMatch = `{
  "schema": "cgo.kernel-truth-seal/1",
  "audit_id": "audit-oracle-0226-match",
  "source": "ebpf",
  "object_sha256": "oracle-object-sha256",
  "retrans_payload_bytes": 327248,
  "retrans_events": 226,
  "qdisc_drop_packets": 226,
  "tc_interface": "e-wan",
  "tc_qdisc": "cake",
  "tc_drops_before": 1,
  "tc_drops_after": 227
}`

const qdiscSealMismatch = `{
  "schema": "cgo.kernel-truth-seal/1",
  "audit_id": "audit-oracle-0226-mismatch",
  "source": "ebpf",
  "object_sha256": "oracle-object-sha256",
  "retrans_payload_bytes": 326800,
  "retrans_events": 225,
  "qdisc_drop_packets": 226,
  "tc_interface": "e-wan",
  "tc_qdisc": "cake",
  "tc_drops_before": 1,
  "tc_drops_after": 227
}`

// Oracle 1 — un compteur NON NUL est recopié verbatim dans le JSON, et
// match=true quand retrans_events == (tc_drops_after - tc_drops_before).
func TestQdiscDropFaithfulCopyNonZero(t *testing.T) {
	dir := t.TempDir()
	in := qdiscSeal(t, dir, "seal-match.json", qdiscSealMatch)
	out := filepath.Join(dir, "report-match.json")
	if code := runConfrontOut(out, in); code != 0 {
		t.Fatalf("runConfrontOut: code=%d, attendu 0", code)
	}
	doc := qdiscDoc(t, out)
	k := qdiscObj(t, doc, "kernel_truth")
	if got := qdiscNum(t, k, "qdisc_drop_packets"); got != 226 {
		t.Fatalf("qdisc_drop_packets copié=%v, attendu 226 (coercition locale)", got)
	}
	if got := qdiscNum(t, k, "retrans_events"); got != 226 {
		t.Fatalf("retrans_events copié=%v, attendu 226", got)
	}
	tc := qdiscObj(t, doc, "tc_s")
	if got := qdiscNum(t, tc, "drops_delta"); got != 226 {
		t.Fatalf("drops_delta=%v, attendu 226", got)
	}
	c := qdiscObj(t, doc, "confrontation")
	if c["match"] != true {
		t.Fatalf("match=%#v, attendu true", c["match"])
	}
}

// Oracle hostile — retrans != delta mais qdisc_drop_packets non nul : le
// verdict doit être false ET le compteur de drops reste recopié. Cet oracle
// prouve que le compteur de drops ne pilote jamais le verdict.
func TestConfrontMatchIndependentOfQdiscDrop(t *testing.T) {
	dir := t.TempDir()
	in := qdiscSeal(t, dir, "seal-mismatch.json", qdiscSealMismatch)
	out := filepath.Join(dir, "report-mismatch.json")
	if code := runConfrontOut(out, in); code != 0 {
		t.Fatalf("runConfrontOut: code=%d, attendu 0", code)
	}
	doc := qdiscDoc(t, out)
	k := qdiscObj(t, doc, "kernel_truth")
	if got := qdiscNum(t, k, "qdisc_drop_packets"); got != 226 {
		t.Fatalf("qdisc_drop_packets copié=%v, attendu 226", got)
	}
	c := qdiscObj(t, doc, "confrontation")
	if c["match"] != false {
		t.Fatalf("match=%#v, attendu false (retrans 225 != delta 226)", c["match"])
	}
	if got := qdiscNum(t, c, "ebpf_retrans_events"); got != 225 {
		t.Fatalf("ebpf_retrans_events=%v, attendu 225", got)
	}
	if got := qdiscNum(t, c, "tc_drops_delta"); got != 226 {
		t.Fatalf("tc_drops_delta=%v, attendu 226", got)
	}
}
