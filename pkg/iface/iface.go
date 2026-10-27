// Package iface — débit d'interface observé par les compteurs noyau.
// stdlib seul : /proc/net/dev sur Linux, un fichier de même format ailleurs
// (les tests Windows passent par ReadCounters avec un fichier temporaire).
package iface

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const procNetDev = "/proc/net/dev"

// Counters — compteurs cumulés d'une interface. Octets et pertes seulement ;
// les paquets ne sont pas exposés, personne ne les affiche.
type Counters struct {
	RXBytes uint64
	RXDrops uint64
	TXBytes uint64
	TXDrops uint64
}

// Sample — un échantillon de débit, champs additifs du frame live.
// Valid=false invalide tous les champs numériques : un zéro n'est jamais
// une mesure dans ce contrat.
type Sample struct {
	Name     string  `json:"iface_name"`
	Source   string  `json:"iface_source"`
	RXMbps   float64 `json:"iface_rx_mbps"`
	TXMbps   float64 `json:"iface_tx_mbps"`
	RXDrops  uint64  `json:"iface_rx_drops"`
	TXDrops  uint64  `json:"iface_tx_drops"`
	WindowMs float64 `json:"iface_window_ms"`
	SampleTS int64   `json:"iface_sample_ts"`
	Valid    bool    `json:"iface_valid"`
	Gap      string  `json:"iface_gap"`
}

// Reader — la couture que l'adaptateur SNMP/RouterOS (phase 2) implémentera.
type Reader interface {
	Read() Sample
	Latest() Sample
}

// NormalizeSource — provenance déclarée, jamais devinée. Toute valeur inconnue
// retombe sur "unknown" ; l'appelant journalise la valeur refusée.
func NormalizeSource(s string) string {
	switch s {
	case "twin", "prod", "unknown":
		return s
	}
	return "unknown"
}

// ReadCounters — parse /proc/net/dev (ou un fichier de même format).
// Deux lignes d'en-tête, puis une ligne par interface. Une ligne tronquée est
// une erreur, pas un saut : lire un fichier partiel produirait une mesure fausse.
func ReadCounters(path string) (map[string]Counters, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]Counters)
	for i, line := range strings.Split(string(raw), "\n") {
		if i < 2 || strings.TrimSpace(line) == "" {
			continue
		}
		colon := strings.LastIndexByte(line, ':')
		if colon < 0 {
			return nil, fmt.Errorf("ligne %d : séparateur ':' absent", i+1)
		}
		name := strings.TrimSpace(line[:colon])
		f := strings.Fields(line[colon+1:])
		if len(f) < 16 {
			return nil, fmt.Errorf("ligne %d (%s) : %d champs, 16 attendus", i+1, name, len(f))
		}
		rx, err := strconv.ParseUint(f[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ligne %d (%s) : rx_bytes : %w", i+1, name, err)
		}
		rxd, err := strconv.ParseUint(f[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ligne %d (%s) : rx_drops : %w", i+1, name, err)
		}
		tx, err := strconv.ParseUint(f[8], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ligne %d (%s) : tx_bytes : %w", i+1, name, err)
		}
		txd, err := strconv.ParseUint(f[11], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ligne %d (%s) : tx_drops : %w", i+1, name, err)
		}
		out[name] = Counters{RXBytes: rx, RXDrops: rxd, TXBytes: tx, TXDrops: txd}
	}
	return out, nil
}
