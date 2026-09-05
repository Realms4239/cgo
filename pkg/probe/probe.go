// Package probe — sondes (ping, petits objets, transfert de masse).
// Toutes les sondes prennent leurs dépendances en interfaces pour tester partout;
// l'exécution réelle se fait sur la VM.
package probe

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PingSample is one completed ping round trip.
type PingSample struct{ RTTms float64 }

// Ping lance un ping (n paquets) et rend les RTT analysés en ms.
// Args par OS : `-c/-i` sur Unix, `-n` sur Windows (1/s, sans -i) —
// `ping -c` exige l'admin sur Windows et rendait l'audit inopérant en mode
// observation. Timeout = durée attendue + 5 s.
func Ping(ctx context.Context, target string, count int, intervalMs int) ([]PingSample, error) {
	args := []string{"-c", strconv.Itoa(count), "-i", fmt.Sprintf("%.2f", float64(intervalMs)/1000), target}
	timeout := time.Duration(count*intervalMs+5000) * time.Millisecond
	if runtime.GOOS == "windows" {
		args = []string{"-n", strconv.Itoa(count), target}
		timeout = time.Duration(count*1000+5000) * time.Millisecond
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, "ping", args...)
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		return nil, fmt.Errorf("ping %s: %w", target, err)
	}
	return parsePing(string(out)), nil
}

// parsePing extrait les valeurs `time=XX.X ms` — tolérant EN + FR
// (`temps=`, Windows francophone : le contexte DSI est francophone) et aux
// locales Ubuntu VM. `<1ms` → moitié d'intervalle (0,5 : convention
// documentée, moins biaisée que l'omission qui surestimerait les agrégats).
func parsePing(out string) []PingSample {
	var s []PingSample
	for _, line := range strings.Split(out, "\n") {
		if v, ok := parsePingLine(line); ok {
			s = append(s, PingSample{RTTms: v})
		}
	}
	sort.Slice(s, func(i, j int) bool { return s[i].RTTms < s[j].RTTms })
	return s
}

func parsePingLine(line string) (float64, bool) {
	norm := strings.Replace(line, "temps=", "time=", 1)
	norm = strings.Replace(norm, "temps<", "time<", 1)
	lt := false
	i := strings.Index(norm, "time=")
	if i < 0 {
		i = strings.Index(norm, "time<")
		if i < 0 {
			return 0, false
		}
		lt = true
	}
	rest := norm[i+5:]
	end := strings.IndexAny(rest, " m")
	if end <= 0 {
		return 0, false
	}
	v, err := strconv.ParseFloat(rest[:end], 64)
	if err != nil {
		return 0, false
	}
	if lt {
		return v / 2, true
	}
	return v, true
}

// smallMarkedTransport — transport HTTP dont les connexions portent DSCP EF
// (Linux uniquement, via dialSmallMarked) : le petit objet vit la classe
// temps réel que l'AQM doit protéger. Hors Linux : transport standard, sonde
// best-effort (documenté — la mesure EF n'est interprétable que sur le banc).
// SmallClient — client HTTP du petit objet : transport marqué DSCP EF
// (Linux/banc). Client partagé, pas de fuite par sonde.
func SmallClient() *http.Client {
	once.Do(func() { markedClient = &http.Client{Timeout: 2 * time.Second, Transport: smallMarkedTransport()} })
	return markedClient
}

// SmallClientTimeout — variante à timeout calibré : le client partagé fixe
// 2 s, mortel sur VSAT (16 Ko à travers netem 600 ms + perte : complétion
// > 2 s → G2 échoue par construction, pas par le lien). Le banc connaît le
// RTT du profil : max(2 s, 4×delay) donne au satellite la marge qu'il
// mérite sans dépayser les profils rapides. Rend un client NEUF : à
// réserver au banc, pas à l'audit (le client partagé y reste).
func SmallClientTimeout(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout, Transport: smallMarkedTransport()}
}

var (
	once         sync.Once
	markedClient *http.Client
)

func smallMarkedTransport() *http.Transport {
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialSmallMarked(ctx, addr)
	}
	return base
}

// SmallObject times one HTTP GET completion in ms.
func SmallObject(ctx context.Context, client *http.Client, url string) (float64, error) {
	t0 := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if err != nil {
		return 0, err
	}
	_ = n
	return float64(time.Since(t0).Microseconds()) / 1000.0, nil
}
