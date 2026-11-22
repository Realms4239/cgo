package api

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/capabilities"
)

// TestSchemaEndCorpusParams — les contrôles du Labo viennent du registre réel :
// canaux, sens, distribution et CC dynamique (bbr3 seulement si le noyau
// l'expose). Aucune valeur n'est fabriquée, aucun libellé inventé.
func TestSchemaEndCorpusParams(t *testing.T) {
	// withCC — stub du noyau, rendu à la fin ; le registre est relu à chaque
	// appel pour refléter la capacité réellement annoncée.
	withCC := func(cc []string, fn func()) {
		restore := capabilities.StubHost(cc, "6.8.0")
		fn()
		restore()
	}

	withCC([]string{"cubic", "bbr"}, func() {
		channel, ok := paramByKey("link_channel")
		if !ok {
			t.Fatal("link_channel absent du registre")
		}
		if !slices.Equal(channel.Enum, []string{"statique", "empirique"}) {
			t.Fatalf("link_channel.enum = %v", channel.Enum)
		}
		if channel.Default != "statique" {
			t.Fatalf("link_channel.default = %v", channel.Default)
		}

		direction, ok := paramByKey("link_direction")
		if !ok {
			t.Fatal("link_direction absent du registre")
		}
		if !slices.Equal(direction.Enum, []string{"up", "down"}) {
			t.Fatalf("link_direction.enum = %v", direction.Enum)
		}
		if direction.Default != "up" {
			t.Fatalf("link_direction.default = %v", direction.Default)
		}

		if _, ok := paramByKey("dist_name"); !ok {
			t.Fatal("dist_name absent du registre")
		}

		state, ok := paramByKey("bbr3_state")
		if !ok {
			t.Fatal("bbr3_state absent du registre")
		}
		const sentence = "BBRv3 non détecté — CUBIC + BBRv1 sur noyau 6.8, détection bbr3 prête"
		if state.Default != sentence {
			t.Fatalf("bbr3_state = %q, attendu %q", state.Default, sentence)
		}

		burst, ok := paramByKey("burst_cc")
		if !ok {
			t.Fatal("burst_cc absent du registre")
		}
		if slices.Contains(burst.Enum, "bbr3") {
			t.Fatalf("burst_cc.enum expose bbr3 sans noyau v3 : %v", burst.Enum)
		}
	})

	withCC([]string{"cubic", "bbr", "bbr3"}, func() {
		burst, ok := paramByKey("burst_cc")
		if !ok {
			t.Fatal("burst_cc absent du registre")
		}
		if !slices.Contains(burst.Enum, "bbr3") {
			t.Fatalf("burst_cc.enum n'expose pas bbr3 pourtant détecté : %v", burst.Enum)
		}
	})
}

// TestBurstAcceptsDetectedBBRv3 — le schéma et la validation du burst lisent
// la même source : si bbr3 est annoncé, POST /api/burst doit l'accepter.
func TestBurstAcceptsDetectedBBRv3(t *testing.T) {
	withDevNoAuth(t)
	called := ""
	srv := httptest.NewServer(New(Deps{BurstFn: func(cc string, _ int) error {
		called = cc
		return nil
	}}))
	defer srv.Close()

	restore := capabilities.StubHost([]string{"cubic", "bbr", "bbr3"}, "6.8.0")
	defer restore()

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/burst", strings.NewReader(`{"cc":"bbr3","seconds":4}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Confirm-Action", "true")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d (bbr3 détecté devrait être accepté)", resp.StatusCode)
	}
	if called != "bbr3" {
		t.Fatalf("burst appelé avec %q, attendu bbr3", called)
	}
}
