package api

import (
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSONEncodesBeforeWriting(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, map[string]any{"ok": true})
	if w.Code != http.StatusOK {
		t.Fatalf("valid value: code = %d, want 200", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Fatal("valid value produced no JSON body")
	}
}

func TestWriteJSONFailsClosedOnEncodeError(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, map[string]any{"bad": math.NaN()})
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("unencodable value: code = %d, want 500", w.Code)
	}
}
