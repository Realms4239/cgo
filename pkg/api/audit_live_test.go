package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Realms4239/cgo/pkg/audit"
	"github.com/Realms4239/cgo/pkg/metrics"
)

type auditLiveResp struct {
	Running        bool    `json:"running"`
	Phase          string  `json:"phase"`
	ElapsedSec     int     `json:"elapsed_sec"`
	TotalSec       int     `json:"total_sec"`
	PhaseRemainSec int     `json:"phase_remain_sec"`
	LastRTTms      float64 `json:"last_rtt_ms"`
	LastSmallMs    float64 `json:"last_small_ms"`
	Samples        struct {
		RTT   []float64 `json:"rtt_ms"`
		Small []float64 `json:"small_ms"`
	} `json:"samples"`
	Timeline []audit.TimelinePoint `json:"timeline"`
	DeadlineMs int                `json:"deadline_ms"`
}

func getJSON(t *testing.T, url string, out any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s : %d", url, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
}

func auditRows(path string) int {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	return bytes.Count(bytes.TrimRight(b, "\r\n"), []byte("\n"))
}

// T5 — l'audit est observable : /api/audit/live expose la phase et les
// échantillons mesurés, /api/audit/stop annule sans écrire de ligne partielle
// et journalise un arrêt volontaire distinct d'une panne.
func TestAuditLiveAndStop(t *testing.T) {
	site := "live-stop-" + time.Now().Format("150405.000000")
	h := New(Deps{
		SinksFn: func() audit.Sinks {
			return audit.Sinks{Mode: "external", Target: "127.0.0.1"}
		},
		AuditDepsFn: func() audit.Deps {
			return audit.Deps{
				Ping:  func(_ context.Context, _ string, _ int) []float64 { return []float64{42, 43, 44, 45, 46} },
				Small: func(context.Context) (float64, error) { return 12.5, nil },
				BulkLoad: func(context.Context, string, string, int) audit.LoadOutcome {
					return audit.LoadOutcome{Bytes: 1_000_000, PerFlow: []uint64{1_000_000}, Elapsed: time.Second}
				},
			}
		},
	})
	srv := httptest.NewServer(h)
	defer srv.Close()
	waitAuditIdle(t, srv)
	csvBefore := auditRows("data/link_audit.csv")

	body, _ := json.Marshal(map[string]any{
		"site": site, "link_type": "fiber", "duration": 10, "direction": "up", "flows": 4,
	})
	resp, err := http.Post(srv.URL+"/api/audit/start", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("start : %d", resp.StatusCode)
	}
	t.Cleanup(func() { _, _ = http.Post(srv.URL+"/api/audit/stop", "application/json", nil) })

	var live auditLiveResp
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		getJSON(t, srv.URL+"/api/audit/live", &live)
		if live.Running && live.LastRTTms > 0 && len(live.Samples.RTT) > 0 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !live.Running || live.Phase == "" || live.LastRTTms == 0 {
		t.Fatalf("live sans état mesuré : %+v", live)
	}
	if live.TotalSec != 10 || live.ElapsedSec < 0 || live.PhaseRemainSec < 0 {
		t.Fatalf("bornes live incohérentes : %+v", live)
	}
	if len(live.Samples.RTT) == 0 || len(live.Samples.RTT) > 32 || len(live.Samples.Small) > 32 {
		t.Fatalf("échantillons live hors bornes : %+v", live.Samples)
	}
	if live.LastSmallMs != 12.5 {
		t.Fatalf("last_small_ms = %v, want 12.5", live.LastSmallMs)
	}
	if live.DeadlineMs != metrics.CAPDeadlineMs {
		t.Fatalf("deadline_ms = %d, want %d", live.DeadlineMs, metrics.CAPDeadlineMs)
	}
	// La timeline couvre tout le run (une entrée par seconde écoulée), pas la
	// fenêtre de 32 sondes : sans ça, la marche repos → charge disparaît.
	if len(live.Timeline) == 0 {
		t.Fatal("timeline live vide alors que des sondes ont répondu")
	}
	if len(live.Timeline) > live.TotalSec+5 {
		t.Fatalf("timeline = %d points pour %d s (plafond total_sec+5)", len(live.Timeline), live.TotalSec)
	}
	for i, pt := range live.Timeline {
		if pt.T != i {
			t.Fatalf("timeline non contiguë : index %d porte t=%d", i, pt.T)
		}
		if pt.RTTP50 != nil && *pt.RTTP50 == 0 {
			t.Fatalf("une sonde muette est tracée comme 0 ms : %+v", pt)
		}
	}

	stopResp, err := http.Post(srv.URL+"/api/audit/stop", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	var stopped struct {
		Stopped bool `json:"stopped"`
	}
	if err := json.NewDecoder(stopResp.Body).Decode(&stopped); err != nil {
		t.Fatal(err)
	}
	stopResp.Body.Close()
	if !stopped.Stopped {
		t.Fatal("stop n'a pas annulé l'audit en cours")
	}

	deadline = time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		live = auditLiveResp{}
		getJSON(t, srv.URL+"/api/audit/live", &live)
		if !live.Running {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if live.Running {
		t.Fatal("live annonce encore l'audit après stop")
	}
	if live.Phase != "" || live.LastRTTms != 0 {
		t.Fatalf("live arrêté ne doit porter aucun zéro mesuré : %+v", live)
	}

	var ev struct {
		Events []struct {
			Kind string `json:"kind"`
			Msg  string `json:"msg"`
		} `json:"events"`
	}
	getJSON(t, srv.URL+"/api/events", &ev)
	annulled := false
	for _, e := range ev.Events {
		if e.Kind == "audit" && strings.Contains(e.Msg, site) && strings.Contains(e.Msg, "annulé par l'opérateur") {
			annulled = true
		}
	}
	if !annulled {
		t.Fatalf("aucun événement « annulé par l'opérateur » pour %s", site)
	}
	if got := auditRows("data/link_audit.csv"); got != csvBefore {
		t.Fatalf("l'annulation a écrit une ligne CSV partielle : %d → %d", csvBefore, got)
	}
}

// TestAuditStartClearsStaleReport — le rapport en mémoire appartient à l'audit
// précédent : /start l'efface avant de lancer le nouveau, sinon un run annulé
// ou en échec afficherait l'ancien rapport comme son résultat.
func TestAuditStartClearsStaleReport(t *testing.T) {
	site := "stale-report-" + time.Now().Format("150405.000000")
	h := New(Deps{
		SinksFn: func() audit.Sinks {
			return audit.Sinks{Mode: "external", Target: "127.0.0.1"}
		},
		AuditDepsFn: func() audit.Deps {
			return audit.Deps{
				Ping:  func(_ context.Context, _ string, _ int) []float64 { return []float64{42, 43, 44} },
				Small: func(context.Context) (float64, error) { return 12.5, nil },
				// Bloque jusqu'à l'annulation : le run reste vivant le temps
				// de lire le statut, sans course avec la fin de l'audit.
				BulkLoad: func(ctx context.Context, _, _ string, _ int) audit.LoadOutcome {
					<-ctx.Done()
					return audit.LoadOutcome{}
				},
			}
		},
	})
	srv := httptest.NewServer(h)
	defer srv.Close()
	waitAuditIdle(t, srv)

	lastAuditReport = &audit.Report{AuditID: "stale-report-" + site}
	t.Cleanup(func() { lastAuditReport = nil })

	body, _ := json.Marshal(map[string]any{
		"site": site, "link_type": "fiber", "duration": 10, "direction": "up", "flows": 4,
	})
	resp, err := http.Post(srv.URL+"/api/audit/start", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("start : %d", resp.StatusCode)
	}
	t.Cleanup(func() { _, _ = http.Post(srv.URL+"/api/audit/stop", "application/json", nil) })

	var st struct {
		Running bool          `json:"running"`
		Report  *audit.Report `json:"report"`
	}
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		st.Report = nil
		getJSON(t, srv.URL+"/api/audit/status", &st)
		if st.Running {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !st.Running {
		t.Fatal("l'audit ne démarre pas")
	}
	if st.Report != nil {
		t.Fatalf("l'ancien rapport survit au démarrage : %+v", st.Report)
	}

	if _, err := http.Post(srv.URL+"/api/audit/stop", "application/json", nil); err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		st.Report = nil
		getJSON(t, srv.URL+"/api/audit/status", &st)
		if !st.Running {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if st.Running {
		t.Fatal("l'audit tourne encore après stop")
	}
	if st.Report != nil {
		t.Fatalf("un run annulé rend un rapport : %+v", st.Report)
	}
}
