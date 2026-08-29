package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// ARG.md pivot: the tool is an edge-shaping CONTROL, not just an observatory —
// the DSI applies the AQM to the local gateway and watches the wall react.
func TestShapeControl(t *testing.T) {
	var mu sync.Mutex
	var calls []string
	h := New(Deps{
		ShapeFn: func(qdisc string, capMbps float64) error {
			mu.Lock()
			defer mu.Unlock()
			calls = append(calls, qdisc)
			if capMbps == 13 {
				return &tcError{msg: "tc exited 1"}
			}
			return nil
		},
	})
	srv := httptest.NewServer(h)
	defer srv.Close()

	post := func(body string) (*http.Response, string) {
		resp, err := http.Post(srv.URL+"/api/shape", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp, string(b)
	}

	resp, _ := post(`{"qdisc":"hfsc","capacity_mbps":20}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid qdisc: %d want 400", resp.StatusCode)
	}

	resp, _ = post(`{"qdisc":"cake"}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing capacity: %d want 400", resp.StatusCode)
	}

	resp, body := post(`{"qdisc":"cake","capacity_mbps":13}`)
	if resp.StatusCode != http.StatusInternalServerError || !strings.Contains(body, "tc exited 1") {
		t.Fatalf("shapefn error: %d %s", resp.StatusCode, body)
	}

	resp, body = post(`{"qdisc":"cake","capacity_mbps":20}`)
	if resp.StatusCode != http.StatusOK || !strings.Contains(body, `"qdisc":"cake"`) || !strings.Contains(body, `"capacity_mbps":20`) {
		t.Fatalf("apply: %d %s", resp.StatusCode, body)
	}

	get, err := http.Get(srv.URL + "/api/shape")
	if err != nil {
		t.Fatal(err)
	}
	defer get.Body.Close()
	var state struct {
		Qdisc string  `json:"qdisc"`
		Cap   float64 `json:"capacity_mbps"`
	}
	_ = json.NewDecoder(get.Body).Decode(&state)
	if state.Qdisc != "cake" || state.Cap != 20 {
		t.Fatalf("state: %+v", state)
	}

	resp, _ = post(`{"qdisc":"none","capacity_mbps":20}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("none: %d", resp.StatusCode)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(calls) < 2 || calls[len(calls)-1] != "none" {
		t.Fatalf("calls: %v", calls)
	}
}

func TestProfilesList(t *testing.T) {
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/profiles")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || !strings.Contains(string(body), `"P1"`) || !strings.Contains(string(body), "capacity_mbps") {
		t.Fatalf("profiles: %d %s", resp.StatusCode, body)
	}
}

type tcError struct{ msg string }

func (e *tcError) Error() string { return e.msg }
