package api

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTranslateUsesBestGoodput — la suggestion CLI doit dériver du goodput de
// la MEILLEURE cellule du profil demandé (P2|cake 18 Mbit/s → 16mbit), pas de
// groups[0] (trié par profil : P1 50 Mbit/s → 45mbit).
func TestTranslateUsesBestGoodput(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()

	runDir := filepath.Join(dir, "data", "runs", "run-x")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatal(err)
	}
	csv := header18 + "\n" +
		"run-x,1,P1,pfifo_fast,bbr,1,20,25,5,100,100,50,0,0,0,0,0,valid\n" +
		"run-x,2,P2,cake,bbr,1,100,114,14,200,98,18,81,0,117288,12.1,0,valid\n" +
		"run-x,3,P2,pfifo_fast,bbr,1,111,168,57,300,60,19,443,5,641464,66.4,0,valid\n"
	if err := os.WriteFile(filepath.Join(runDir, "aqm_eval.csv"), []byte(csv), 0644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/api/hardware/translate?profile=P2", nil)
	rec := httptest.NewRecorder()
	HandleTranslate(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, `"available":true`) {
		t.Fatalf("translate unavailable: %s", body)
	}
	if !strings.Contains(body, "16mbit") {
		t.Fatalf("cli bandwidth must derive from best cell goodput 18 (→16mbit), got: %s", body)
	}
	if strings.Contains(body, "45mbit") {
		t.Fatalf("cli bandwidth leaks groups[0] (P1 50 Mbit/s): %s", body)
	}
	if !strings.Contains(body, "cake") {
		t.Fatalf("verdict must name best qdisc cake: %s", body)
	}
}
