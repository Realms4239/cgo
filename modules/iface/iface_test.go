package iface

import (
	"os"
	"path/filepath"
	"testing"
)

const fixture = `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: 1000 10 0 2 0 0 0 0 2000 20 0 3 0 0 0 0
  eth0: 123456789012 100 0 4 0 0 0 0 987654321098 200 0 5 0 0 0 0
 eth:0: 18446744073709551615 7 0 8 0 0 0 0 9 1 0 2 0 0 0 0
`

func writeFile(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "dev")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestReadCountersParses(t *testing.T) {
	got, err := ReadCounters(writeFile(t, fixture))
	if err != nil {
		t.Fatal(err)
	}
	if c := got["lo"]; c.RXBytes != 1000 || c.RXDrops != 2 || c.TXBytes != 2000 || c.TXDrops != 3 {
		t.Fatalf("lo = %+v", c)
	}
	if c := got["eth0"]; c.RXBytes != 123456789012 || c.TXDrops != 5 {
		t.Fatalf("eth0 = %+v", c)
	}
	// le nom peut contenir ':' : on coupe au DERNIER séparateur
	if c := got["eth:0"]; c.RXBytes != 18446744073709551615 {
		t.Fatalf("eth:0 = %+v", c)
	}
}

func TestReadCountersRejectsShortLine(t *testing.T) {
	body := "a\nb\n eth0: 1 2 3\n"
	if _, err := ReadCounters(writeFile(t, body)); err == nil {
		t.Fatal("ligne courte acceptée : une ligne tronquée doit faire échouer la lecture")
	}
}

func TestReadCountersMissingFile(t *testing.T) {
	if _, err := ReadCounters(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("fichier absent accepté")
	}
}

func TestNormalizeSource(t *testing.T) {
	for in, want := range map[string]string{
		"": "unknown", "twin": "twin", "prod": "prod", "production": "unknown",
	} {
		if got := NormalizeSource(in); got != want {
			t.Fatalf("%q -> %q, want %q", in, got, want)
		}
	}
}
