package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestProfileImportFullColumns — CSV étendu : up asymétrique + rafales
// Gilbert-Elliott aux positions 5..9 ; réponse GET /api/profiles les
// expose (sinon l'import UI/API écrit des champs que personne ne lit).
func TestProfileImportFullColumns(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()

	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	csv := "PXF,20,100,15,0.5,5,0.01,0.001,200\n"
	req, _ := http.NewRequest("POST", srv.URL+"/api/profile/import", strings.NewReader(csv))
	req.Header.Set("Content-Type", "text/csv")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("import csv étendu: %d", resp.StatusCode)
	}

	resp2, err := http.Get(srv.URL + "/api/profiles")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	b, _ := io.ReadAll(resp2.Body)
	s := string(b)
	for _, want := range []string{`"capacity_up_mbps":5`, `"loss_burst_p":0.01`, `"loss_burst_r":0.001`, `"loss_burst_h":200`} {
		if !strings.Contains(s, want) {
			t.Fatalf("GET /api/profiles sans %s dans %s", want, s)
		}
	}
}
