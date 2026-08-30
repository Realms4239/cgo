package campagne

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Realms4239/cgo/pkg/model"
)

// Writer ajoute des lignes au aqm_eval.csv gelé d'un run et gèle le
// manifest. G5 (pas de lignes événement dupliquées) est appliqué ici, à l'unique
// chemin d'écriture.
type Writer struct {
	Dir        string
	seen       map[string]bool
	fh         *os.File
	dupes      int
	quarantine []quarantineEntry
}

type quarantineEntry struct {
	EventID int    `json:"event_id"`
	Profile string `json:"profile"`
	Qdisc   string `json:"qdisc"`
	CC      string `json:"cc"`
	Status  string `json:"gate_status"`
}

func OpenRun(dir string) (*Writer, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	w := &Writer{Dir: dir, seen: map[string]bool{}}
	path := filepath.Join(dir, "aqm_eval.csv")
	exists := false
	if st, err := os.Stat(path); err == nil && st.Size() > 0 {
		exists = true
		// remplir seen pour la reprise (G5) — lire les lignes existantes
		if data, err := os.ReadFile(path); err == nil {
			lines := splitLines(string(data))
			for i, ln := range lines {
				if i == 0 {
					continue // en-tête
				}
				if ln == "" {
					continue
				}
				parts := splitCSV(ln)
				if len(parts) < 2 {
					continue
				}
				w.seen[parts[0]+"/"+parts[1]] = true
			}
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	w.fh = f
	if !exists {
		hdr := joinRow(model.AQMEvalHeader)
		if _, err := f.WriteString(hdr + "\n"); err != nil {
			return nil, err
		}
	}
	return w, nil
}

func splitLines(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == '\n' {
			out = append(out, cur)
			cur = ""
		} else if r != '\r' {
			cur += string(r)
		}
	}
	out = append(out, cur)
	return out
}

func splitCSV(s string) []string { return splitOn(s, ',') }

func splitOn(s string, sep rune) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == sep {
			out = append(out, cur)
			cur = ""
		} else {
			cur += string(r)
		}
	}
	out = append(out, cur)
	return out
}

func joinRow(cols []string) string {
	out := ""
	for i, c := range cols {
		if i > 0 {
			out += ","
		}
		out += c
	}
	return out
}

// Append écrit une ligne ; les paires (run,event) dupliquées sont rejetées (G5).
func (w *Writer) Append(ev model.Event) error {
	key := fmt.Sprintf("%s/%d", ev.RunID, ev.EventID)
	if w.seen[key] {
		w.dupes++
		return fmt.Errorf("duplicate event row %s", key)
	}
	w.seen[key] = true
	if ev.GateStatus != model.GatePass {
		w.quarantine = append(w.quarantine, quarantineEntry{
			EventID: ev.EventID, Profile: string(ev.Profile), Qdisc: string(ev.Qdisc), CC: string(ev.CC), Status: ev.GateStatus,
		})
	}
	row := fmt.Sprintf("%s,%d,%s,%s,%s,%d,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%d,%d,%d,%.2f,%.1f,%s",
		ev.RunID, ev.EventID, ev.Profile, ev.Qdisc, ev.CC, ev.Repetition,
		ev.RTTp50Ms, ev.RTTp95Ms, ev.QDIPctMs, ev.Smallp95Ms, ev.DeadlineOKPct,
		ev.BulkGoodputMbps, ev.Drops, ev.Retransmissions, ev.WastedBytes,
		ev.CostARPerH, ev.CPUPct, ev.GateStatus)
	_, err := fmt.Fprintln(w.fh, row)
	return err
}

// Freeze ferme le CSV et écrit manifest.json avec le SHA-256 de chaque
// fichier, plus quarantine.json pour les événements degraded/invalid.
func (w *Writer) Freeze(cfgHash string) error {
	if err := w.fh.Close(); err != nil {
		return err
	}
	// quarantaine
	if len(w.quarantine) > 0 {
		qb, _ := json.MarshalIndent(w.quarantine, "", "  ")
		_ = os.WriteFile(filepath.Join(w.Dir, "quarantine.json"), qb, 0644)
	}
	type entry struct{ File, Sha256 string }
	var files []entry
	err := filepath.Walk(w.Dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Base(p) == "manifest.json" {
			return err
		}
		// inclure quarantine.json dans le hash du manifest
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(b)
		files = append(files, entry{
			File:   filepath.Base(p),
			Sha256: hex.EncodeToString(sum[:]),
		})
		return nil
	})
	if err != nil {
		return err
	}
	doc, _ := json.MarshalIndent(map[string]any{"config_sha256": cfgHash, "files": files}, "", " ")
	return os.WriteFile(filepath.Join(w.Dir, "manifest.json"), doc, 0o644)
}
