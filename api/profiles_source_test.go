package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestAPICitedProfilesSource — /api/profiles sert les deux cibles d'émulation
// avec leur provenance, sans les marquer importées.
func TestAPICitedProfilesSource(t *testing.T) {
	withDevNoAuth(t)
	chdirTemp(t) // profile.Load lit data/profiles.json relatif au CWD
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/profiles")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)

	var doc struct {
		Profiles []struct {
			ID       string `json:"id"`
			Source   string `json:"source"`
			Imported bool   `json:"imported"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("GET /api/profiles: %v (%s)", err, b)
	}

	seen := map[string]bool{}
	for _, p := range doc.Profiles {
		if p.ID != "L1" && p.ID != "V1" {
			continue
		}
		seen[p.ID] = true
		if strings.TrimSpace(p.Source) == "" {
			t.Fatalf("%s: source absente de /api/profiles", p.ID)
		}
		if p.Imported {
			t.Fatalf("%s: marqué importé à tort", p.ID)
		}
	}
	for _, id := range []string{"L1", "V1"} {
		if !seen[id] {
			t.Fatalf("%s absent de /api/profiles: %s", id, b)
		}
	}
}
