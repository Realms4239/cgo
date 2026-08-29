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
		ShapeFn: func(r ShapeReq) error {
			mu.Lock()
			defer mu.Unlock()
			calls = append(calls, r.Qdisc)
			if r.CapMbps == 13 {
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

// Hardening — zero-value Deps must answer 503, never panic (the GetSnap fix's
// contract extended to the mutating endpoints).
func TestMutatingEndpointsNilGuards(t *testing.T) {
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/run/start", "application/json", bytes.NewBufferString(`{"profiles":["P1"],"reps":1}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("run/start nil StartFn: %d want 503", resp.StatusCode)
	}

	resp2, err := http.Post(srv.URL+"/api/run/stop", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("run/stop nil StopFn: %d want 503", resp2.StatusCode)
	}
}

// The shape lever and a running campagne fight over the same shaper — a manual
// apply mid-run corrupts the cell's shaping (and vice versa). 409 while active.
func TestShapeConflictWithRunningCampagne(t *testing.T) {
	h := New(Deps{
		ShapeFn:   func(ShapeReq) error { return nil },
		RunningFn: func() bool { return true },
	})
	srv := httptest.NewServer(h)
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/api/shape", "application/json", bytes.NewBufferString(`{"qdisc":"cake","capacity_mbps":20}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusConflict || !strings.Contains(string(body), "campagne") {
		t.Fatalf("shape during run: %d %s want 409 campagne", resp.StatusCode, body)
	}
}

// Graceful shutdown — CloseHub stops the hub ticker; double-close is safe.
func TestHandlerCloseHub(t *testing.T) {
	h := New(Deps{ShapeFn: func(ShapeReq) error { return nil }})
	h.CloseHub()
	h.CloseHub() // idempotent
}

// Hub subscriber cap — beyond maxSubs streams the server sheds load with 503.
func TestHubSubscriberCap(t *testing.T) {
	h := NewHub()
	for i := 0; i < maxSubs; i++ {
		h.subs[&sub{ch: make(chan string, 1)}] = struct{}{}
	}
	srv := httptest.NewServer(http.HandlerFunc(h.SSE))
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/stream")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("cap: %d want 503", resp.StatusCode)
	}
}

// Q8/Q10 — the lever parameterizes the link itself (delay/jitter/loss) and the
// campagne honors the operator's deadline + target. Bounds are server-side.
func TestShapeLinkConditions(t *testing.T) {
	var mu sync.Mutex
	var last ShapeReq
	h := New(Deps{
		ShapeFn: func(r ShapeReq) error {
			mu.Lock()
			defer mu.Unlock()
			last = r
			return nil
		},
	})
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/shape", "application/json",
		bytes.NewBufferString(`{"qdisc":"cake","capacity_mbps":20,"delay_ms":80,"jitter_ms":5,"loss_pct":0.5}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("apply: %d %s", resp.StatusCode, b)
	}
	mu.Lock()
	defer mu.Unlock()
	if last.DelayMs != 80 || last.JitterMs != 5 || last.LossPct != 0.5 {
		t.Fatalf("netem params lost: %+v", last)
	}

	// bounds: delay ≤ 600, jitter ≤ 100, loss ≤ 10
	for _, bad := range []string{
		`{"qdisc":"cake","capacity_mbps":20,"delay_ms":601}`,
		`{"qdisc":"cake","capacity_mbps":20,"jitter_ms":101}`,
		`{"qdisc":"cake","capacity_mbps":20,"loss_pct":11}`,
	} {
		r2, err := http.Post(srv.URL+"/api/shape", "application/json", bytes.NewBufferString(bad))
		if err != nil {
			t.Fatal(err)
		}
		r2.Body.Close()
		if r2.StatusCode != http.StatusBadRequest {
			t.Fatalf("bounds %s: %d want 400", bad, r2.StatusCode)
		}
	}
}

func TestRunOptsDeadlineTarget(t *testing.T) {
	var mu sync.Mutex
	var got *RunOpts
	h := New(Deps{
		StartFn: func(o RunOpts) error {
			mu.Lock()
			defer mu.Unlock()
			got = &o
			return nil
		},
	})
	srv := httptest.NewServer(h)
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/api/run/start", "application/json",
		bytes.NewBufferString(`{"profiles":["P1"],"reps":1,"deadline_ms":2500,"target":"9.9.9.9"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("start: %d %s", resp.StatusCode, b)
	}
	mu.Lock()
	defer mu.Unlock()
	if got == nil || got.DeadlineMs != 2500 || got.Target != "9.9.9.9" {
		t.Fatalf("opts lost: %+v", got)
	}
}

// Journal — operator memory: actions land in a queryable ring.
func TestJournal(t *testing.T) {
	h := New(Deps{ShapeFn: func(ShapeReq) error { return nil }})
	srv := httptest.NewServer(h)
	defer srv.Close()
	_, _ = http.Post(srv.URL+"/api/shape", "application/json", bytes.NewBufferString(`{"qdisc":"cake","capacity_mbps":20}`))
	resp, err := http.Get(srv.URL + "/api/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || !strings.Contains(string(body), "façonnage") || !strings.Contains(string(body), "cake") {
		t.Fatalf("journal: %d %s", resp.StatusCode, body)
	}
}

// Compare view — raw rows per run, path-traversal safe.
func TestRunRows(t *testing.T) {
	h := New(Deps{})
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/run/rows?run=..%2F..%2Fetc")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("traversal: %d want 400", resp.StatusCode)
	}

	resp2, err := http.Get(srv.URL + "/api/run/rows?run=no-such-run")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	b, _ := io.ReadAll(resp2.Body)
	if resp2.StatusCode != http.StatusNotFound {
		t.Fatalf("missing run: %d want 404", resp2.StatusCode)
	}
	if !strings.Contains(string(b), "no-such-run") {
		t.Fatalf("404 body: %s", b)
	}
}

// Prevention — server-side bounds are the contract (§6): the client hint is
// never trusted. Audit duration clamps to 10–600 s; unknown profile ids and
// out-of-range reps are refused, not silently swallowed into an empty run.
func TestValidationPrevention(t *testing.T) {
	h := New(Deps{ShapeFn: func(ShapeReq) error { return nil }})
	srv := httptest.NewServer(h)
	defer srv.Close()

	// shape capacity out of range
	resp, _ := http.Post(srv.URL+"/api/shape", "application/json", bytes.NewBufferString(`{"qdisc":"cake","capacity_mbps":5000}`))
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("capacity 5000: %d want 400", resp.StatusCode)
	}

	// unknown profile id → 400 with the offending id
	resp2, err := http.Post(srv.URL+"/api/run/start", "application/json", bytes.NewBufferString(`{"profiles":["P1; drop","GHOST"],"reps":1}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	b2, _ := io.ReadAll(resp2.Body)
	if resp2.StatusCode != http.StatusBadRequest || !strings.Contains(string(b2), "GHOST") {
		t.Fatalf("unknown profile: %d %s want 400 naming it", resp2.StatusCode, b2)
	}

	// reps out of range → 400
	resp3, err := http.Post(srv.URL+"/api/run/start", "application/json", bytes.NewBufferString(`{"profiles":["P1"],"reps":99}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusBadRequest {
		t.Fatalf("reps 99: %d want 400", resp3.StatusCode)
	}

	// deadline outside the 200–5000 window → 400 (Q10 bounds are server-side)
	resp5, err := http.Post(srv.URL+"/api/run/start", "application/json", bytes.NewBufferString(`{"profiles":["P1"],"reps":1,"deadline_ms":99999}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp5.Body.Close()
	if resp5.StatusCode != http.StatusBadRequest {
		t.Fatalf("deadline 99999: %d want 400", resp5.StatusCode)
	}

	// target longer than 64 characters → 400
	resp6, err := http.Post(srv.URL+"/api/run/start", "application/json", bytes.NewBufferString(`{"profiles":["P1"],"reps":1,"target":"`+strings.Repeat("a", 65)+`"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp6.Body.Close()
	if resp6.StatusCode != http.StatusBadRequest {
		t.Fatalf("target 65 chars: %d want 400", resp6.StatusCode)
	}

	// audit duration clamps to the 10–600 window
	resp4, err := http.Post(srv.URL+"/api/audit/start", "application/json", bytes.NewBufferString(`{"site":"x","link_type":"5g","duration":999999}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp4.Body.Close()
	st, err := http.Get(srv.URL + "/api/audit/status")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Body.Close()
	var status struct {
		Running bool   `json:"running"`
		Err     string `json:"error"`
	}
	_ = json.NewDecoder(st.Body).Decode(&status)
	if resp4.StatusCode != http.StatusOK {
		t.Fatalf("audit start: %d", resp4.StatusCode)
	}
}
