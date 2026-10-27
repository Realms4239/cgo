package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Realms4239/cgo/pkg/audit"
)

// TestCrossSiteGuard — une page tierce ne peut pas déclencher une mutation
// par un POST navigateur (CSRF) ; la lecture et les clients sans en-tête
// (curl, CLI, tests) restent inchangés.
func TestCrossSiteGuard(t *testing.T) {
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()

	post := func(crossSite bool) int {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/audit/stop", nil)
		if err != nil {
			t.Fatal(err)
		}
		if crossSite {
			req.Header.Set("Sec-Fetch-Site", "cross-site")
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if code := post(true); code != http.StatusForbidden {
		t.Fatalf("POST cross-site : %d, want 403", code)
	}
	if code := post(false); code != http.StatusOK {
		t.Fatalf("POST sans en-tête : %d, want 200", code)
	}

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/health", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET cross-site : %d, want 200 (lecture non bloquée)", resp.StatusCode)
	}
}

// TestReadAuditReportRejectsUnsafeID — la lecture applique la même garde que
// l'écriture : un id malformé ne résout pas silencieusement un autre fichier.
func TestReadAuditReportRejectsUnsafeID(t *testing.T) {
	for _, id := range []string{"..", ".", "../x", `..\x`, "a/b"} {
		if _, err := readAuditReport(id); err == nil {
			t.Fatalf("id %q accepté par la lecture", id)
		}
	}
}

// TestTruncateUTF8 — la coupe à 80 octets ne fend jamais un caractère.
func TestTruncateUTF8(t *testing.T) {
	s := strings.Repeat("é", 50) // 100 octets, 50 runes
	got := truncateUTF8(s, 80)
	if len(got) != 80 || !utf8.ValidString(got) {
		t.Fatalf("coupe invalide : %d octets, valid=%v", len(got), utf8.ValidString(got))
	}
	if want := strings.Repeat("é", 40); got != want {
		t.Fatalf("coupe = %q, want %q", got, want)
	}
	if got := truncateUTF8("abc", 80); got != "abc" {
		t.Fatalf("chaîne courte modifiée : %q", got)
	}
}

// TestAuditStartTrimsTarget — une cible entourée d'espaces ne doit pas
// partir telle quelle vers le ping : elle est assainie avant le banc par
// défaut, et l'événement d'ouverture le prouve.
func TestAuditStartTrimsTarget(t *testing.T) {
	t.Setenv("CGO_BULK_ADDR", "")
	t.Setenv("CGO_SMALL_URL", "")
	h := New(Deps{
		SinksFn: func() audit.Sinks {
			return audit.Sinks{Mode: "external", Target: "127.0.0.1"}
		},
		AuditDepsFn: func() audit.Deps {
			return audit.Deps{
				Ping:  func(context.Context, string, int) []float64 { return []float64{42} },
				Small: func(context.Context) (float64, error) { return 1, nil },
			}
		},
	})
	srv := httptest.NewServer(h)
	defer srv.Close()
	waitAuditIdle(t, srv)

	body, _ := json.Marshal(map[string]any{
		"site": "trim-" + time.Now().Format("150405.000000"), "link_type": "fiber",
		"duration": 10, "direction": "up", "target": "  10.30.0.2  ",
	})
	resp, err := http.Post(srv.URL+"/api/audit/start", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("start : %d", resp.StatusCode)
	}

	var ev struct {
		Events []struct {
			Msg string `json:"msg"`
		} `json:"events"`
	}
	getJSON(t, srv.URL+"/api/events", &ev)
	found := false
	for _, e := range ev.Events {
		if strings.Contains(e.Msg, "cible 10.30.0.2,") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("cible non assainie dans l'événement : %+v", ev.Events)
	}

	if _, err := http.Post(srv.URL+"/api/audit/stop", "application/json", nil); err != nil {
		t.Fatal(err)
	}
	waitAuditIdle(t, srv)
}

// TestAuditPanicRecovered — une dépendance qui panique ne tue pas le serveur :
// la goroutine détachée journalise la panne, efface le rapport de ce run et
// rend auditRunning=false pour l'audit suivant.
func TestAuditPanicRecovered(t *testing.T) {
	t.Setenv("CGO_BULK_ADDR", "")
	t.Setenv("CGO_SMALL_URL", "")
	site := "panic-" + time.Now().Format("150405.000000")
	h := New(Deps{
		SinksFn: func() audit.Sinks {
			return audit.Sinks{Mode: "external", Target: "127.0.0.1"}
		},
		AuditDepsFn: func() audit.Deps {
			return audit.Deps{
				Ping:  func(context.Context, string, int) []float64 { panic("panne simulée") },
				Small: func(context.Context) (float64, error) { return 1, nil },
			}
		},
	})
	srv := httptest.NewServer(h)
	defer srv.Close()
	waitAuditIdle(t, srv)

	body, _ := json.Marshal(map[string]any{
		"site": site, "link_type": "fiber", "duration": 10, "direction": "up",
	})
	resp, err := http.Post(srv.URL+"/api/audit/start", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("start : %d", resp.StatusCode)
	}

	var st struct {
		Running bool          `json:"running"`
		Report  *audit.Report `json:"report"`
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		st = struct {
			Running bool          `json:"running"`
			Report  *audit.Report `json:"report"`
		}{}
		getJSON(t, srv.URL+"/api/audit/status", &st)
		if !st.Running {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if st.Running {
		t.Fatal("l'audit paniqué reste « running »")
	}
	if st.Report != nil {
		t.Fatalf("un rapport survit à la panne : %+v", st.Report)
	}

	// Le serveur répond encore : la panique est restée dans la goroutine.
	health, err := http.Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatal(err)
	}
	health.Body.Close()
	if health.StatusCode != http.StatusOK {
		t.Fatalf("serveur inutilisable après panne : %d", health.StatusCode)
	}

	var ev struct {
		Events []struct {
			Kind string `json:"kind"`
			Msg  string `json:"msg"`
		} `json:"events"`
	}
	getJSON(t, srv.URL+"/api/events", &ev)
	found := false
	for _, e := range ev.Events {
		if strings.Contains(e.Msg, "panne interne") && strings.Contains(e.Msg, site) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("panne non journalisée : %+v", ev.Events)
	}
}
