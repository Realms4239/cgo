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

// TestDecodeStatsRootCountersLiveShape — forme REELLE de `tc -j -s` sur le
// banc sadcnwp (Ubuntu 24.04, iproute2 ≥ 6.x) : les compteurs sont à la
// racine de la ligne, pas dans un objet "stats". Capture du 2026-11-16 :
// `dropped 19588` en texte correspondait à `"drops":19588` en JSON, mais
// l'ancien decodeur ne lisait que `stats.dropped` -> 0 partout (drops,
// wasted_bytes, coût). Régression : une cellule cake à 19 588 drops doit
// les rapporter exactement.
func TestDecodeStatsRootCountersLiveShape(t *testing.T) {
	out := []byte(`[{"kind":"cake","handle":"1:","root":true,"refcnt":5,
      "bytes":348051838,"packets":230591,"drops":19588,"overlimits":700137,
      "requeues":0,"backlog":7570,"qlen":5,"memory_used":677618}]`)
	stats, err := decodeStats(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 {
		t.Fatalf("want 1 row, got %d: %+v", len(stats), stats)
	}
	want := Stats{Kind: "cake", Handle: "1:", Bytes: 348051838,
		Packets: 230591, Drops: 19588, Overlimits: 700137, Backlog: 7570}
	if stats[0] != want {
		t.Fatalf("root counters lost: got %+v, want %+v", stats[0], want)
	}
}

// Les deux formes peuvent cohabiter : la racine l'emporte quand elle est
// renseignée, l'objet imbriqué ne comble que les champs absents.
func TestDecodeStatsRootWinsOverNestedFallback(t *testing.T) {
	out := []byte(`[{"kind":"cake","handle":"1:","drops":19588,
      "stats":{"bytes":777,"packets":9,"dropped":3,"overlimits":0,"backlog":0}}]`)
	stats, err := decodeStats(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 || stats[0].Drops != 19588 || stats[0].Bytes != 777 {
		t.Fatalf("root should win, nested should fill: %+v", stats)
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
