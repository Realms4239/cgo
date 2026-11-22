package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeCalibrationFixture — sidecar honnête : son hash scelle les octets du
// .dist réelLEMENT présent. Rien n'est synthétisé par le serveur.
func writeCalibrationFixture(t *testing.T, dir, name string, dist []byte, samples int, p50 float64) {
	t.Helper()
	sum := sha256.Sum256(dist)
	raw, err := json.Marshal(map[string]any{
		"sha256":  hex.EncodeToString(sum[:]),
		"samples": samples,
		"p50_ms":  p50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".dist"), dist, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCalibrations — GET /api/calibrations ne liste que les sidecars vérifiés :
// un .dist orphelin et un sidecar au hash divergent sont écartés, sans PASS
// inventé ni valeur fabriquée.
func TestCalibrations(t *testing.T) {
	withDevNoAuth(t)
	dir := t.TempDir()
	t.Setenv("CGO_CALIBRATION_DIR", dir)

	dist := []byte("100\n110\n120\n")
	writeCalibrationFixture(t, dir, "P2", dist, 35, 110)

	// entrée invalide 1 : .dist orphelin, aucun sidecar.
	if err := os.WriteFile(filepath.Join(dir, "orphan.dist"), []byte("1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// entrée invalide 2 : sidecar dont le hash ne scelle pas le .dist.
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte(`{"sha256":"deadbeef","samples":10,"p50_ms":50}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.dist"), []byte("1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/api/calibrations")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var got struct {
		Calibrations []Calibration `json:"calibrations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got.Calibrations) != 1 {
		t.Fatalf("calibrations = %+v, attendu une seule entrée vérifiée", got.Calibrations)
	}
	sum := sha256.Sum256(dist)
	want := Calibration{
		Name: "P2", SHA256: hex.EncodeToString(sum[:]), Samples: 35, P50Ms: 110, Verified: true,
	}
	if got.Calibrations[0] != want {
		t.Fatalf("calibration = %+v, attendu %+v", got.Calibrations[0], want)
	}
}

// TestCalibrationsEmpty — un répertoire vide est une réponse honnête, pas une
// erreur : [] avec 200, jamais une entrée fabriquée pour meubler la liste.
func TestCalibrationsEmpty(t *testing.T) {
	withDevNoAuth(t)
	t.Setenv("CGO_CALIBRATION_DIR", t.TempDir())

	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/api/calibrations")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"calibrations":[]`) {
		t.Fatalf("corps = %s, attendu calibrations vide", strings.TrimSpace(string(body)))
	}
}
