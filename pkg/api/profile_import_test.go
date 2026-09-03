package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
)

// Import de profil CSV — une ligne sans en-tête : id,capacity_mbps,delay_ms,jitter_ms,loss_pct
// Chdir vers un dossier temp : profile.Import persiste data/profiles.json relatif au CWD — ne jamais polluer le dépôt.
func TestProfileImportCSV(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	req, err := http.NewRequest("POST", srv.URL+"/api/profile/import", strings.NewReader("PX,20,100,15,0.5\n"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "text/csv")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("csv import status = %d, want 200", res.StatusCode)
	}
	p, ok := model.Profiles["PX"]
	if !ok {
		t.Fatal("PX not registered after csv import")
	}
	if p.CapacityMbps != 20 || p.DelayMs != 100 || p.JitterMs != 15 || p.LossPct != 0.5 {
		t.Fatalf("bad csv values: %+v", p)
	}
}
