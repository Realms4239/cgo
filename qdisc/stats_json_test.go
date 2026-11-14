package qdisc

import (
	"strings"
	"testing"
)

// Fixtures réalistes `tc -j -s qdisc show` (iproute2 ≥ 5.x) : deux
// orthographes de drops, deux formes de handle.
func TestDecodeStatsMixedRows(t *testing.T) {
	out := []byte(`[
  {"kind":"netem","handle":"1:","stats":{"bytes":9596383791,"packets":6351949,"dropped":281,"overlimits":0,"backlog":0}},
  {"kind":"tbf","handle":{"major":10,"minor":1},"stats":{"bytes":8000000,"packets":5300,"drops":12,"overlimits":5,"backlog":1500}},
  {"kind":"cake","handle":{"major":10,"minor":2},"stats":{"bytes":777,"packets":9,"dropped":3,"overlimits":0,"backlog":0}}
]`)
	stats, err := decodeStats(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 3 {
		t.Fatalf("want 3 rows, got %d: %+v", len(stats), stats)
	}
	if stats[0].Kind != "netem" || stats[0].Handle != "1:" || stats[0].Drops != 281 {
		t.Fatalf("netem row: %+v", stats[0])
	}
	if stats[1].Handle != "a:1" || stats[1].Drops != 12 {
		t.Fatalf("tbf row (handle objet, drops): %+v", stats[1])
	}
	if stats[2].Handle != "a:2" || stats[2].Drops != 3 {
		t.Fatalf("cake row: %+v", stats[2])
	}
}

// pfifo_fast est invisible dans `tc -j qdisc` (tc-pfifo_fast(8)) : tableau
// vide attendu, zéro ligne, pas d'erreur.
func TestDecodeStatsEmptyArrayIsZeroRows(t *testing.T) {
	stats, err := decodeStats([]byte(`[]`))
	if err != nil {
		t.Fatalf("empty array must not be an error: %v", err)
	}
	if len(stats) != 0 {
		t.Fatalf("want zero rows, got %+v", stats)
	}
}

// JSON malformé : erreur nommée, jamais un retour silencieux au texte.
func TestDecodeStatsGarbageFailsClosed(t *testing.T) {
	_, err := decodeStats([]byte("qdisc netem 1: root refcnt 129\n Sent 42 bytes 1 pkt"))
	if err == nil {
		t.Fatal("garbage text must not decode")
	}
	if !strings.Contains(err.Error(), "JSON") {
		t.Fatalf("error must name the parse problem: %v", err)
	}
}

// Une ligne sans objet stats (qdisc sans compteur) = compteurs à zéro,
// pas d'erreur.
func TestDecodeStatsRowWithoutStats(t *testing.T) {
	stats, err := decodeStats([]byte(`[{"kind":"ingress","handle":"ffff:"}]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 || stats[0].Kind != "ingress" || stats[0].Bytes != 0 {
		t.Fatalf("row without stats: %+v (err %v)", stats, err)
	}
}
