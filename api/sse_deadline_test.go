package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type deadlineSpyResponseWriter struct {
	http.ResponseWriter
	deadlineCleared bool
}

func (s *deadlineSpyResponseWriter) Unwrap() http.ResponseWriter { return s.ResponseWriter }

func (s *deadlineSpyResponseWriter) SetWriteDeadline(t time.Time) error {
	if t.IsZero() {
		s.deadlineCleared = true
	}
	return nil
}

func (s *deadlineSpyResponseWriter) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func TestSSEClearsWriteDeadline(t *testing.T) {
	h := NewHub()
	// Contexte déjà annulé : SSE sort immédiatement de sa boucle, on ne teste
	// que le dégagement de la deadline en tête de handler.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/stream", nil).WithContext(ctx)
	c := &deadlineSpyResponseWriter{ResponseWriter: w}
	h.SSE(c, r)
	if !c.deadlineCleared {
		t.Error("Hub.SSE must clear the write deadline for long-lived streams")
	}
}
