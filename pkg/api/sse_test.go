package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type snap struct {
	Phase    string  `json:"phase"`
	Profile  string  `json:"profile"`
	RTTp50Ms float64 `json:"rtt_p50_ms"`
}

// Publisher runs at exactly 10 Hz; the assertion covers fan-out + delta.
func TestSSECadenceAndDelta(t *testing.T) {
	cur := snap{Phase: "baseline", Profile: "P1", RTTp50Ms: 20}
	h := NewHub()

	mux := http.NewServeMux()
	mux.Handle("GET /api/stream", http.HandlerFunc(h.SSE))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1300*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/stream", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	go func() {
		tk := time.NewTicker(time.Second / frameHz)
		defer tk.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tk.C:
				cur.RTTp50Ms++
				h.Publish(cur) // profile unchanged → omitted after first frame
			}
		}
	}()

	data, _ := io.ReadAll(resp.Body)
	frames := strings.Count(string(data), "\ndata:")
	profiles := strings.Count(string(data), `"profile"`)
	if frames < 8 || frames > 18 {
		t.Fatalf("cadence off: %d frames in ~1.2 s @10Hz", frames)
	}
	if profiles != 1 {
		t.Fatalf("structural delta broken: profile appeared %d times", profiles)
	}
}
