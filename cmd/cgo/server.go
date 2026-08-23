package main

import (
	"context"
	"net/http"
)

// runServer starts the dashboard+API server. M0 scaffold: /api/health only;
// routes fill in at M1.4. Kept here so main.go stays command-routing only.
func runServer(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	srv := &http.Server{Addr: addr, Handler: mux}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	select {
	case <-ctx.Done():
		return srv.Close()
	case err := <-errCh:
		return err
	}
}
