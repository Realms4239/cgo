package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Realms4239/cgo/audit"
)

// waitAuditIdle — le statut d'audit est global au serveur : un audit lancé
// par un test voisin ferait 409 au suivant. On attend la fin des audits
// résiduels avant de démarrer le nôtre (et après l'avoir lancé). Plafond
// 50 s : la deadline interne d'un audit 10 s est durée+30, et un connect
// Windows vers une adresse non routée peut consommer ses retransmissions
// SYN (~21 s) avant l'échec.
func waitAuditIdle(t *testing.T, srv *httptest.Server) {
	t.Helper()
	deadline := time.Now().Add(50 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(srv.URL + "/api/audit/status")
		if err != nil {
			t.Fatal(err)
		}
		var st struct {
			Running bool `json:"running"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&st)
		resp.Body.Close()
		if !st.Running {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("audit résiduel toujours actif après 20 s")
}

// auditStartEvent — la ligne « démarré » du journal opérateur pour le site
// demandé : le POST audit/start revient avant la fin de l'audit, mais la
// ligne d'ouverture est enregistrée de façon synchrone.
func auditStartEvent(t *testing.T, srv *httptest.Server, site string) string {
	t.Helper()
	resp, err := http.Get(srv.URL + "/api/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out struct {
		Events []struct {
			Kind string `json:"kind"`
			Msg  string `json:"msg"`
		} `json:"events"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	for _, e := range out.Events {
		if e.Kind == "audit" && strings.Contains(e.Msg, site) && strings.Contains(e.Msg, "démarré") {
			return e.Msg
		}
	}
	t.Fatalf("aucune ligne « audit démarré » pour %s dans le journal", site)
	return ""
}

// TestAuditStartSinksBenchAware — les champs vides suivent le banc DÉTECTÉ
// (SinksFn stub, hors réseau), pas un 10.200.0.1 en dur : sur le banc durci
// dualns, cible, petit objet et charge portent le trio 10.30.0.2 — les
// QUATRE sauts sur le chemin mesuré.
func TestAuditStartSinksBenchAware(t *testing.T) {
	t.Setenv("CGO_ALLOW_PRIVATE_TARGETS", "1") // opt-in banc : cibles 10.30.0.2
	h := New(Deps{SinksFn: func() audit.Sinks {
		return audit.Sinks{Mode: "dualns", Target: "10.30.0.2", SmallURL: "http://10.30.0.2:8081/small", BulkAddr: "10.30.0.2:5201"}
	}})
	srv := httptest.NewServer(h)
	defer srv.Close()
	waitAuditIdle(t, srv)

	resp, err := http.Post(srv.URL+"/api/audit/start", "application/json",
		bytes.NewBufferString(`{"site":"puits-dualns","link_type":"5g","duration":10}`))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("audit/start dualns: %d, want 200", resp.StatusCode)
	}
	msg := auditStartEvent(t, srv, "puits-dualns")
	for _, want := range []string{"dualns", "10.30.0.2", "http://10.30.0.2:8081/small", "10.30.0.2:5201"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("ligne démarré sans %q : %s", want, msg)
		}
	}
	// l'audit réel tourne en goroutine sur le statut GLOBAL — le laisser
	// finir, sinon le test voisin qui démarre un audit hérite d'un 409.
	waitAuditIdle(t, srv)
}

// TestAuditStartSinksExternalNoImplicitLoad — en external, le petit objet
// reste VIDE (jamais de charge implicite — un A+ sans charge est un mensonge)
// et la charge ne vient que de CGO_BULK_ADDR ; la ligne du journal le dit.
func TestAuditStartSinksExternalNoImplicitLoad(t *testing.T) {
	t.Setenv("CGO_BULK_ADDR", "127.0.0.1:1")
	t.Setenv("CGO_ALLOW_PRIVATE_TARGETS", "1") // loopback RFC1918-adjacent : opt-in requis
	h := New(Deps{SinksFn: func() audit.Sinks {
		return audit.Sinks{Mode: "external", Target: "8.8.8.8"}
	}})
	srv := httptest.NewServer(h)
	defer srv.Close()
	waitAuditIdle(t, srv)

	// bulk explicite = même adresse que l'env : l'audit échoue vite (port
	// refusé) et ne retient pas le statut global.
	resp, err := http.Post(srv.URL+"/api/audit/start", "application/json",
		bytes.NewBufferString(`{"site":"puits-external","link_type":"fiber","duration":10,"bulk_addr":"127.0.0.1:1"}`))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("audit/start external: %d, want 200", resp.StatusCode)
	}
	msg := auditStartEvent(t, srv, "puits-external")
	if !strings.Contains(msg, "external") || !strings.Contains(msg, "8.8.8.8") {
		t.Fatalf("ligne démarré doit nommer le banc et la cible : %s", msg)
	}
	if !strings.Contains(msg, "127.0.0.1:1") {
		t.Fatalf("la charge d'opérateur doit passer (env ou explicite) : %s", msg)
	}
	if strings.Contains(msg, "8081") {
		t.Fatalf("external : jamais de charge implicite, ligne dit un small : %s", msg)
	}
	// même hygiène : l'audit résiduel (bulk refusé, il échoue vite) doit
	// avoir libéré le statut global avant le test suivant.
	waitAuditIdle(t, srv)
}

// TestAuditStartValidation400 — branches 400 épinglées (M5) : direction
// invalide, flux hors bornes, cible trop longue. Ces refus passent AVANT la
// porte « audit déjà en cours » (la validation ne dépend pas de l'état).
func TestAuditStartValidation400(t *testing.T) {
	srv := httptest.NewServer(New(Deps{}))
	defer srv.Close()

	post := func(body string) int {
		resp, err := http.Post(srv.URL+"/api/audit/start", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		return resp.StatusCode
	}

	if code := post(`{"site":"x","direction":"sideways"}`); code != http.StatusBadRequest {
		t.Fatalf("direction invalide: %d, want 400", code)
	}
	if code := post(`{"site":"x","flows":9}`); code != http.StatusBadRequest {
		t.Fatalf("flows 9: %d, want 400", code)
	}
	if code := post(`{"site":"x","target":"` + strings.Repeat("a", 65) + `"}`); code != http.StatusBadRequest {
		t.Fatalf("cible 65 caractères: %d, want 400", code)
	}
}
