package api

import (
	"encoding/json"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
)

func TestHealthExposesTelemetrySource(t *testing.T) {
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct {
		Telemetry struct {
			Source string `json:"source"`
			Reason string `json:"reason"`
		} `json:"telemetry"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Telemetry.Source != "tcpinfo" {
		t.Fatalf("source = %q, want tcpinfo on this host", body.Telemetry.Source)
	}
	if body.Telemetry.Reason == "" {
		t.Fatal("fallback without a reason")
	}
	// Sur Windows il n'y a ni eBPF ni échantillonneur interface : /api/health
	// doit le dire, pas revendiquer une sonde noyau absente.
	if runtime.GOOS == "windows" {
		if body.Telemetry.Source != "tcpinfo" || !strings.Contains(body.Telemetry.Reason, "non-Linux") {
			t.Fatalf("télémétrie Windows mal étiquetée : %+v", body.Telemetry)
		}
	}
}
