package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/metrics"
)

// Sync TUI ↔ web — la fixture partagée audit-verdict-sync.json doit être
// octet pour octet identique des deux côtés, et le TUI doit rendre le mot
// du moteur (« Lien stable »), jamais un mot de remplacement.

const (
	fixtureGoWeb = "../../web/frontend/src/lib/audit-verdict-sync.json"
	fixtureGo    = "testdata/audit-verdict-sync.json"
)

type syncFixture struct {
	OperatorVerdict struct {
		Level    string `json:"level"`
		Word     string `json:"word"`
		Sentence string `json:"sentence"`
	} `json:"operator_verdict"`
}

func readOrFail(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("fixture attendue absente (%s): %v", p, err)
	}
	return b
}

func TestAuditVerdictSyncFixtureByteIdentity(t *testing.T) {
	wb := readOrFail(t, fixtureGoWeb)
	gb := readOrFail(t, fixtureGo)
	if string(wb) != string(gb) {
		t.Fatalf("fixture désynchronisée entre web (%d octets) et Go (%d octets)", len(wb), len(gb))
	}
}

func TestAuditVerdictSyncFixtureCarriesEngineWord(t *testing.T) {
	var fx syncFixture
	if err := json.Unmarshal(readOrFail(t, fixtureGo), &fx); err != nil {
		t.Fatal(err)
	}
	if fx.OperatorVerdict.Word != "Lien stable" {
		t.Fatalf("mot du moteur attendu « Lien stable », vu « %s »", fx.OperatorVerdict.Word)
	}
	if fx.OperatorVerdict.Level != "green" {
		t.Fatalf("niveau attendu green, vu %s", fx.OperatorVerdict.Level)
	}
}

func TestAuditVerdictSyncTUIRendersEngineWord(t *testing.T) {
	var fx syncFixture
	if err := json.Unmarshal(readOrFail(t, fixtureGo), &fx); err != nil {
		t.Fatal(err)
	}
	v := auditView{OK: true, Verdict: metrics.Verdict{Level: metrics.VerdictLevel(fx.OperatorVerdict.Level), Word: fx.OperatorVerdict.Word, Sentence: fx.OperatorVerdict.Sentence}}
	out := verdictLine(v)
	if !strings.Contains(out, fx.OperatorVerdict.Word) {
		t.Fatalf("ligne TUI sans le mot du moteur: %q", out)
	}
	if !strings.Contains(out, fx.OperatorVerdict.Sentence) {
		t.Fatalf("ligne TUI sans la phrase du moteur: %q", out)
	}
}
