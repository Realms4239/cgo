package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/audit"
)

// TestAuthorizeOwner — oracle direct de la décision d'accès à une ressource
// possédée. Un rapport historique sans propriétaire reste lisible une fois
// authentifié (côté ServeHTTP), jamais modifiable par un simple opérateur.
func TestAuthorizeOwner(t *testing.T) {
	admin := Principal{OperatorID: "aaaaaaaaaaaaaaaa", Role: "admin"}
	op := Principal{OperatorID: "bbbbbbbbbbbbbbbb", Role: "operator"}
	if !authorizeResource(admin, "cccccccccccccccc") {
		t.Error("admin must access any resource")
	}
	if !authorizeResource(op, "bbbbbbbbbbbbbbbb") {
		t.Error("owner must access own resource")
	}
	if authorizeResource(op, "cccccccccccccccc") {
		t.Error("cross-owner access must be refused")
	}
	if !authorizeResource(op, "") {
		t.Error("legacy unowned resource must be readable when authenticated")
	}
}

// TestAuthorizeMutation — oracle hostile : une ressource sans propriétaire
// n'est modifiable que par l'admin ; un opérateur ne modifie que la sienne.
func TestAuthorizeMutation(t *testing.T) {
	admin := Principal{OperatorID: "aaaaaaaaaaaaaaaa", Role: "admin"}
	op := Principal{OperatorID: "bbbbbbbbbbbbbbbb", Role: "operator"}
	if !authorizeMutation(admin, "cccccccccccccccc") || !authorizeMutation(admin, "") {
		t.Error("admin must mutate any resource, owned or legacy")
	}
	if !authorizeMutation(op, "bbbbbbbbbbbbbbbb") {
		t.Error("owner must mutate own resource")
	}
	if authorizeMutation(op, "cccccccccccccccc") {
		t.Error("cross-owner mutation must be refused")
	}
	if authorizeMutation(op, "") {
		t.Error("legacy unowned resource must not be mutable by an operator")
	}
}

// TestReportOwnerIDJSONKey — le contrat de persistance : owner_id est une clé
// JSON du rapport, présente même quand elle est vide (rapport historique).
func TestReportOwnerIDJSONKey(t *testing.T) {
	raw, err := json.Marshal(&audit.Report{AuditID: "audit-1", OwnerID: "bbbbbbbbbbbbbbbb"})
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if got, ok := payload["owner_id"]; !ok || got != "bbbbbbbbbbbbbbbb" {
		t.Fatalf("owner_id = %v (present=%v), want bbbbbbbbbbbbbbbb", got, ok)
	}
	legacy, err := json.Marshal(&audit.Report{AuditID: "audit-2"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(legacy), `"owner_id":""`) {
		t.Fatalf("rapport historique sans owner_id : clé absente de %s", legacy)
	}
}

// TestOwnerGuardsReturnForbidden — les enveloppes traduisent la décision en
// 403. Sans Bearer, principalFrom rend un Principal vide (non-admin) : c'est le
// chemin hostile, sans injecter de faux rôle dans l'application. Les lecteurs
// authentifiés sont, eux, refusés par ServeHTTP avant d'atteindre l'enveloppe.
func TestOwnerGuardsReturnForbidden(t *testing.T) {
	t.Setenv("CGO_API_TOKEN", "token-secret-abc")
	next := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }
	owned := func(*http.Request) string { return "cccccccccccccccc" }
	legacy := func(*http.Request) string { return "" }
	call := func(h http.HandlerFunc, bearer string) int {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/api/audit/report", nil)
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		rec := httptest.NewRecorder()
		h(rec, req)
		return rec.Code
	}
	srv := &Server{}
	if code := call(srv.requireOwner(next, owned), ""); code != http.StatusForbidden {
		t.Errorf("cross-owner read : %d, want 403", code)
	}
	if code := call(srv.requireOwner(next, legacy), ""); code != http.StatusNoContent {
		t.Errorf("legacy read : %d, want 204", code)
	}
	if code := call(srv.requireOwnerMutate(next, owned), ""); code != http.StatusForbidden {
		t.Errorf("cross-owner mutation : %d, want 403", code)
	}
	if code := call(srv.requireOwnerMutate(next, legacy), ""); code != http.StatusForbidden {
		t.Errorf("legacy mutation : %d, want 403", code)
	}
	if code := call(srv.requireOwnerMutate(next, owned), "token-secret-abc"); code != http.StatusNoContent {
		t.Errorf("admin cross-owner mutation : %d, want 204", code)
	}
	if code := call(srv.requireOwnerMutate(next, legacy), "token-secret-abc"); code != http.StatusNoContent {
		t.Errorf("admin legacy mutation : %d, want 204", code)
	}
}
