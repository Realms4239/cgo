package iface

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func devLine(rx, tx, rxd, txd uint64) string {
	return "h1\nh2\n eth0: " + strconv.FormatUint(rx, 10) + " 0 0 " +
		strconv.FormatUint(rxd, 10) + " 0 0 0 0 " +
		strconv.FormatUint(tx, 10) + " 0 0 " +
		strconv.FormatUint(txd, 10) + " 0 0 0 0\n"
}

func writeDev(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func newTestSampler(t *testing.T, now *time.Time) (*Sampler, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dev")
	return newSampler(path, "eth0", "twin", func() time.Time { return *now }), path
}

func TestSamplerWarmupThenRate(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	writeDev(t, path, devLine(1000, 2000, 1, 2))

	first := s.Read()
	if first.Valid || first.Gap != "warmup" {
		t.Fatalf("premier échantillon = %+v, want warmup", first)
	}
	if first.SampleTS != now.UnixMilli() || first.Name != "eth0" || first.Source != "twin" {
		t.Fatalf("warmup = %+v", first)
	}

	now = now.Add(100 * time.Millisecond)
	writeDev(t, path, devLine(1000+125000, 2000+25000, 3, 4))
	got := s.Read()
	if !got.Valid || got.Gap != "" {
		t.Fatalf("échantillon = %+v, want valid", got)
	}
	if !near(got.RXMbps, 10) || !near(got.TXMbps, 2) {
		t.Fatalf("débits = %.3f/%.3f, want 10/2", got.RXMbps, got.TXMbps)
	}
	if got.WindowMs != 100 {
		t.Fatalf("fenêtre = %v, want 100", got.WindowMs)
	}
	if got.SampleTS != now.UnixMilli() {
		t.Fatalf("ts = %d, want %d", got.SampleTS, now.UnixMilli())
	}
	if got.RXDrops != 3 || got.TXDrops != 4 {
		t.Fatalf("pertes = %d/%d", got.RXDrops, got.TXDrops)
	}
}

func TestSamplerResetRebaselines(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	writeDev(t, path, devLine(500000, 500000, 0, 0))
	_ = s.Read()

	now = now.Add(100 * time.Millisecond)
	writeDev(t, path, devLine(10, 10, 0, 0))
	if got := s.Read(); got.Valid || got.Gap != "reset" || got.WindowMs != 100 {
		t.Fatalf("reset = %+v", got)
	}

	now = now.Add(100 * time.Millisecond)
	writeDev(t, path, devLine(10+125000, 10+25000, 0, 0))
	if got := s.Read(); !got.Valid || !near(got.RXMbps, 10) {
		t.Fatalf("reprise = %+v", got)
	}
}

func TestSamplerSlowAndClock(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	writeDev(t, path, devLine(0, 0, 0, 0))
	_ = s.Read()

	now = now.Add(600 * time.Millisecond)
	writeDev(t, path, devLine(1250000, 0, 0, 0))
	if got := s.Read(); got.Valid || got.Gap != "slow" || got.WindowMs != 600 {
		t.Fatalf("slow = %+v", got)
	}

	now = now.Add(100 * time.Millisecond)
	writeDev(t, path, devLine(2500000, 0, 0, 0))
	if got := s.Read(); !got.Valid || !near(got.RXMbps, 100) {
		t.Fatalf("après slow = %+v", got)
	}

	now = now.Add(-50 * time.Millisecond)
	if got := s.Read(); got.Valid || got.Gap != "clock" || got.WindowMs != 0 {
		t.Fatalf("clock = %+v", got)
	}
}

func TestSamplerAbsentThenWarmup(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	writeDev(t, path, "h1\nh2\n lo: 1 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0\n")
	if got := s.Read(); got.Valid || got.Gap != "absent" {
		t.Fatalf("absent = %+v", got)
	}

	writeDev(t, path, devLine(100, 100, 0, 0))
	now = now.Add(100 * time.Millisecond)
	if got := s.Read(); got.Gap != "warmup" {
		t.Fatalf("retour = %+v, want warmup", got)
	}
	now = now.Add(100 * time.Millisecond)
	writeDev(t, path, devLine(100+12500, 100, 0, 0))
	if got := s.Read(); !got.Valid || !near(got.RXMbps, 1) {
		t.Fatalf("reprise = %+v", got)
	}
}

func TestSamplerReadError(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	writeDev(t, path, devLine(0, 0, 0, 0))
	_ = s.Read()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if got := s.Read(); got.Valid || got.Gap != "read" {
		t.Fatalf("read = %+v", got)
	}
}

func TestSamplerLatest(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	if got := s.Latest(); got.Gap != "warmup" || got.Name != "eth0" || got.Source != "twin" {
		t.Fatalf("latest initial = %+v", got)
	}
	writeDev(t, path, devLine(0, 0, 0, 0))
	first := s.Read()
	if got := s.Latest(); got != first {
		t.Fatalf("latest = %+v, want %+v", got, first)
	}
}

func TestSamplerConcurrentReadLatest(t *testing.T) {
	now := time.Unix(1000, 0)
	s, path := newTestSampler(t, &now)
	writeDev(t, path, devLine(0, 0, 0, 0))
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				s.Read()
				s.Latest()
			}
		}()
	}
	wg.Wait()
}

func TestNewNormalizesSource(t *testing.T) {
	s := New("eth0", "production")
	if got := s.Latest().Source; got != "unknown" {
		t.Fatalf("source = %q, want unknown", got)
	}
	if got := s.Latest().Name; got != "eth0" {
		t.Fatalf("name = %q", got)
	}
}
