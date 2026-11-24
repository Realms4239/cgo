package campagne

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Realms4239/cgo/internal/kit"
	"github.com/Realms4239/cgo/model"
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
	// FailedGates — portes en échec (runs futurs ; absent des archives passées).
	FailedGates []string `json:"failed_gates,omitempty"`
}

func OpenRun(dir string) (*Writer, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	w := &Writer{Dir: dir, seen: map[string]bool{}}
	path := filepath.Join(dir, "aqm_eval.csv")
	exists := false
	// Réparation de queue AVANT l'O_APPEND : une ligne tronquée par l'arrêt
	// (dernier octet ≠ '\n') n'a jamais été gelée — l'O_APPEND collerait sinon
	// la nouvelle ligne saine contre les fragments de la déchirée (un CSV
	// poursuivant sur la même ligne physique) et le manifeste hash ensuite des
	// octets jamais mesurés. On tronque après le DERNIER '\n' (0 s'il n'y en
	// a aucun — le fichier repart alors sur un en-tête neuf, exists=false).
	if st, err := os.Stat(path); err == nil && st.Size() > 0 {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reprise: lecture impossible de %s : %w", path, err)
		}
		if len(data) > 0 && data[len(data)-1] != '\n' {
			if i := bytes.LastIndexByte(data, '\n'); i >= 0 {
				if err := os.Truncate(path, int64(i+1)); err != nil {
					return nil, fmt.Errorf("reprise: troncature de %s : %w", path, err)
				}
				data = data[:i+1]
			} else {
				if err := os.Truncate(path, 0); err != nil {
					return nil, fmt.Errorf("reprise: troncature de %s : %w", path, err)
				}
				data = nil
			}
		}
		// remplir seen pour la reprise (G5) — lire les lignes existantes
		if len(data) > 0 {
			exists = true
			lines := splitLines(string(data))
			for i, ln := range lines {
				if i == 0 {
					continue // en-tête
				}
				if ln == "" {
					continue
				}
				parts := splitCSV(ln)
				// identité de cellule : profil/qdisc/cc/sens/rep (colonnes 3-7).
				// La position (event_id) ne sert plus : une reprise sous-matrice
				// renumérote et ferait sauter des cellules jamais gelées.
				// Une ligne coupée en deux (arrêt PC en pleine écriture) peut
				// porter 7+ cellules sans en porter 23 : on exige la largeur
				// complète pour enregistrer l'identité, sinon la reprise
				// « voit » une cellule jamais gelée et G5 la saute à tort.
				// Limite connue : une déchirure APRÈS la cellule 23 (dans
				// gate_status, ~7 octets de fenêtre) se lit complète —
				// identité (colonnes 3-7) déjà saine, fenêtre acceptée.
				if len(parts) < 23 {
					continue
				}
				w.seen[parts[2]+"/"+parts[3]+"/"+parts[4]+"/"+parts[5]+"/"+parts[6]] = true
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

// cellKey — identité d'une cellule indépendante de sa position : une reprise
// en sous-matrice (--qdiscs cake après une pleine) renumérote les event_id ;
// la position ne peut pas servir d'identité. (run_id imposé par le CLI de
// reprise : une nouvelle matrice sans --run-id ne partage jamais le CSV.)
func cellKey(ev model.Event) string {
	return fmt.Sprintf("%s/%s/%s/%s/%d", ev.Profile, ev.Qdisc, ev.CC, ev.Direction, ev.Repetition)
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

// Append écrit une ligne ; les cellules identiques (profil/qdisc/cc/sens/rep,
// position indépendante) sont rejetées (G5).
func (w *Writer) Append(ev model.Event) error {
	key := cellKey(ev)
	if w.seen[key] {
		w.dupes++
		return fmt.Errorf("duplicate event row %s", key)
	}
	w.seen[key] = true
	if ev.GateStatus != model.GatePass {
		w.quarantine = append(w.quarantine, quarantineEntry{
			EventID: ev.EventID, Profile: string(ev.Profile), Qdisc: string(ev.Qdisc), CC: string(ev.CC), Status: ev.GateStatus,
			FailedGates: ev.FailedGates,
		})
	}
	row := fmt.Sprintf("%s,%d,%s,%s,%s,%s,%d,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%d,%d,%d,%.2f,%.1f,%.1f,%.1f,%s",
		ev.RunID, ev.EventID, ev.Profile, ev.Qdisc, ev.CC, ev.Direction, ev.Repetition,
		ev.RTTp50Ms, ev.RTTp95Ms, ev.QDIPctMs, ev.VoIPR, ev.JFI, ev.Smallp95Ms, ev.DeadlineOKPct,
		ev.BulkGoodputMbps, ev.Drops, ev.Retransmissions, ev.WastedBytes,
		ev.CostARPerH, ev.CPUPct, ev.RTTBaseP50Ms, ev.RTTBaseP95Ms, ev.GateStatus)
	_, err := fmt.Fprintln(w.fh, row)
	return err
}

// Freeze ferme le CSV et écrit manifest.json avec le SHA-256 de chaque
// fichier, le SHA-256 de la configuration effective et cette configuration
// elle-même (vérifiable hors ligne). Plus quarantine.json pour les événements
// degraded/invalid. Une configuration absente ou non-JSON est refusée : le
// manifeste ne doit jamais porter un hash factice.
func (w *Writer) Freeze(configJSON []byte) error {
	// la ligne gelée d'abord sur le disque, le manifest ensuite : le
	// manifeste hash le fichier — il doit hash des octets durables.
	syncErr := w.fh.Sync()
	closeErr := w.fh.Close()
	if syncErr != nil {
		return syncErr
	}
	if closeErr != nil {
		return closeErr
	}
	if len(configJSON) == 0 || !json.Valid(configJSON) {
		return fmt.Errorf("freeze: configuration absente ou JSON invalide")
	}
	cfgSum := sha256.Sum256(configJSON)
	cfgHash := hex.EncodeToString(cfgSum[:])
	// quarantaine — écriture insécable : un arrêt en pleine sauvegarde ne
	// laisse jamais un quarantine.json tronqué que le manifeste hash ensuite.
	if len(w.quarantine) > 0 {
		qb, _ := json.MarshalIndent(w.quarantine, "", "  ")
		_ = kit.WriteFileAtomic(filepath.Join(w.Dir, "quarantine.json"), qb, 0644)
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
	// Contrat de topologie : empreinte ADDITIVE des octets déclarés. Sans
	// CGO_TOPOLOGY_CONTRACT, le champ est omis (jamais de hash factice) ;
	// un chemin configuré mais illisible fait échouer le gel.
	topoSha := ""
	if topoPath := os.Getenv("CGO_TOPOLOGY_CONTRACT"); topoPath != "" {
		b, err := os.ReadFile(topoPath)
		if err != nil {
			return fmt.Errorf("freeze: contrat de topologie illisible (%s): %w", topoPath, err)
		}
		sum := sha256.Sum256(b)
		topoSha = hex.EncodeToString(sum[:])
	}
	manifest := map[string]any{
		"config_sha256": cfgHash,
		"config":        string(configJSON),
		"files":         files,
	}
	if topoSha != "" {
		manifest["topology_contract_sha256"] = topoSha
	}
	doc, _ := json.MarshalIndent(manifest, "", " ")
	return kit.WriteFileAtomic(filepath.Join(w.Dir, "manifest.json"), doc, 0o644)
}
