package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/audit"
)

// TestAuditStartWhileRunningReturns409 — un second démarrage pendant un audit
// réellement en cours est refusé par la session, sans toucher au premier run.
func TestAuditStartWhileRunningReturns409(t *testing.T) {
	withDevNoAuth(t)
	release := make(chan struct{})
	deps := ssrfDeps()
	deps.AuditDepsFn = func() audit.Deps {
		return audit.Deps{
			Ping: func(ctx context.Context, _ string, _ int) []float64 {
				select {
				case <-release:
				case <-ctx.Done():
				}
				return []float64{42}
			},
			Small: func(context.Context) (float64, error) { return 1, nil },
		}
	}
	srv := httptest.NewServer(New(deps))
	defer srv.Close()
	waitAuditIdle(t, srv)

	code, body := startAudit(t, srv, `{"site":"conflict-first","link_type":"fiber","duration":10}`)
	if code != http.StatusOK {
		t.Fatalf("premier audit : %d, voulu 200 (corps %q)", code, body)
	}

	code, body = startAudit(t, srv, `{"site":"conflict-second","link_type":"fiber","duration":10}`)
	if code != http.StatusConflict {
		t.Fatalf("second audit pendant le premier : %d, voulu 409 (corps %q)", code, body)
	}
	if !strings.Contains(body, "audit already running") {
		t.Fatalf("409 sans motif explicite : %q", body)
	}

	close(release)
	waitAuditIdle(t, srv)
}
