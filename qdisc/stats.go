package qdisc

import (
	"encoding/json"
	"fmt"
)

// Stats — compteurs cumulés par qdisc, lus depuis `tc -j -s qdisc show`.
type Stats struct {
	Kind       string
	Handle     string
	Bytes      uint64
	Packets    uint64
	Drops      uint64
	Overlimits uint64
	Backlog    uint64
}

// PollStats lit `tc -j -s qdisc show dev <iface>` et décode le JSON iproute2.
// Fonctionne en ns principal (veth-c) et netns (veth-s) par TCRunner.
// pfifo_fast est invisible dans `tc -j qdisc` (tc-pfifo_fast(8)) : zéro ligne
// est attendu pour une cellule pfifo, pas une erreur de parsing.
func PollStats(r TCRunner, iface string) ([]Stats, error) {
	out, err := r.Run("-j", "-s", "qdisc", "show", "dev", iface)
	if err != nil {
		return nil, err
	}
	return decodeStats(out)
}

// tcQdiscRow — sous-ensemble du schéma JSON iproute2 (≥ 5.x) ; champs
// inconnus ignorés.
type tcQdiscRow struct {
	Kind   string          `json:"kind"`
	Handle json.RawMessage `json:"handle"`
	// iproute2 ≥ 6.x (Ubuntu 24.04) place les compteurs à la racine de la
	// ligne, pas dans un objet "stats" imbriqué. Les deux formes sont lues :
	// un schema imbriqué uniquement ramenait tous les compteurs à zéro.
	Bytes      uint64 `json:"bytes"`
	Packets    uint64 `json:"packets"`
	Dropped    uint64 `json:"dropped"`
	Drops      uint64 `json:"drops"`
	Overlimits uint64 `json:"overlimits"`
	Backlog    uint64 `json:"backlog"`
	Stats      *tcStatsJSON `json:"stats"`
}

// withFallback — la forme racine l'emporte ; la forme imbriquée ne comble
// que les champs absents pour rester compatible avec les anciens iproute2.
func (t tcStatsJSON) withFallback(f tcStatsJSON) tcStatsJSON {
	if t.Bytes == 0 {
		t.Bytes = f.Bytes
	}
	if t.Packets == 0 {
		t.Packets = f.Packets
	}
	if t.Dropped == 0 {
		t.Dropped = f.Dropped
	}
	if t.Drops == 0 {
		t.Drops = f.Drops
	}
	if t.Overlimits == 0 {
		t.Overlimits = f.Overlimits
	}
	if t.Backlog == 0 {
		t.Backlog = f.Backlog
	}
	return t
}

type tcStatsJSON struct {
	Bytes      uint64 `json:"bytes"`
	Packets    uint64 `json:"packets"`
	Dropped    uint64 `json:"dropped"`
	Drops      uint64 `json:"drops"` // orthographe alternative selon versions iproute2
	Overlimits uint64 `json:"overlimits"`
	Backlog    uint64 `json:"backlog"`
}

// decodeStats — fail-closed : JSON malformé = erreur nommée, jamais un
// retour silencieux au scrape texte.
func decodeStats(out []byte) ([]Stats, error) {
	var rows []tcQdiscRow
	if err := json.Unmarshal(out, &rows); err != nil {
		return nil, fmt.Errorf("tc -j -s qdisc: JSON illisible: %w", err)
	}
	stats := make([]Stats, 0, len(rows))
	for _, row := range rows {
		s := Stats{Kind: row.Kind, Handle: renderHandle(row.Handle)}
		// Compteurs à la racine (iproute2 réel) ; l'objet imbriqué reste
		// supporté pour les anciennes versions.
		st := tcStatsJSON{
			Bytes: row.Bytes, Packets: row.Packets,
			Dropped: row.Dropped, Drops: row.Drops,
			Overlimits: row.Overlimits, Backlog: row.Backlog,
		}
		if row.Stats != nil {
			st = st.withFallback(*row.Stats)
		}
		s.Bytes = st.Bytes
		s.Packets = st.Packets
		// Les deux orthographes du même compteur sont tolérées.
		s.Drops = st.Dropped
		if st.Drops > s.Drops {
			s.Drops = st.Drops
		}
		s.Overlimits = st.Overlimits
		s.Backlog = st.Backlog
		stats = append(stats, s)
	}
	return stats, nil
}

// renderHandle tolère les deux formes de `tc -j` : chaîne « 1: » ou objet
// {"major":1,"minor":0} rendu « 1:0 ».
func renderHandle(raw json.RawMessage) string {
	var s string
	if len(raw) > 0 && json.Unmarshal(raw, &s) == nil {
		return s
	}
	var h struct {
		Major int `json:"major"`
		Minor int `json:"minor"`
	}
	if len(raw) > 0 && json.Unmarshal(raw, &h) == nil {
		return fmt.Sprintf("%x:%x", h.Major, h.Minor)
	}
	return ""
}

// SumDrops additionne les pertes de tous les qdiscs non racines (delta).
func SumDrops(stats []Stats) uint64 {
	var total uint64
	for _, s := range stats {
		total += s.Drops
	}
	return total
}

// SumBytes rend les octets par le compteur du qdisc feuille.
// Shaper empilé sur la même sortie: netem 1: parent de tbf 10: / cake 10:
// avec fq_codel 20: enfant du tbf. Les mêmes paquets sont comptés à chaque
// layer, so summing inflates goodput 2-3× (P1 80Mbit observed 231).
// Leaf is last in tc output (fq_codel if present, else tbf/cake).
func SumBytes(stats []Stats) uint64 {
	if len(stats) == 0 {
		return 0
	}
	// Feuille seule; somme par couche si l'heuristique se trompe sur un qdisc exotique.
	return stats[len(stats)-1].Bytes
}
