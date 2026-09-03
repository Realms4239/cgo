// Package probe — sondes (ping, petits objets, transfert de masse).
// Toutes les sondes prennent leurs dépendances en interfaces pour tester partout;
// l'exécution réelle se fait sur la VM.
package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// PingSample is one completed ping round trip.
type PingSample struct{ RTTms float64 }

// Ping lance `ping -c n -i interval` et rend les RTT analysés en ms.
// Exécution directe sous contexte (timeout = durée attendue + 5 s) : pas de
// seam injectable — l'interface CmdRunner historique n'était jamais appelée
// (son contrat sans contexte ne pouvait pas porter le timeout).
func Ping(ctx context.Context, target string, count int, intervalMs int) ([]PingSample, error) {
	cctx, cancel := context.WithTimeout(ctx, time.Duration(count*intervalMs+5000)*time.Millisecond)
	defer cancel()
	args := []string{"-c", strconv.Itoa(count), "-i", fmt.Sprintf("%.2f", float64(intervalMs)/1000), target}
	cmd := exec.CommandContext(cctx, "ping", args...)
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		return nil, fmt.Errorf("ping %s: %w", target, err)
	}
	return parsePing(string(out)), nil
}

// parsePing extrait les valeurs `time=XX.X ms` (tolérant aux locales pour
// Ubuntu VM; documented ceiling: non-English ping output on exotic hosts).
func parsePing(out string) []PingSample {
	var s []PingSample
	for _, line := range strings.Split(out, "\n") {
		i := strings.Index(line, "time=")
		if i < 0 {
			continue
		}
		rest := line[i+5:]
		end := strings.IndexAny(rest, " m")
		if end <= 0 {
			continue
		}
		if v, err := strconv.ParseFloat(rest[:end], 64); err == nil {
			s = append(s, PingSample{RTTms: v})
		}
	}
	sort.Slice(s, func(i, j int) bool { return s[i].RTTms < s[j].RTTms })
	return s
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
