package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/Realms4239/cgo/figures"
	"github.com/Realms4239/cgo/results"
)

// corpusCache — memoïsation du bootstrap du corpus. Les gelures sont
// append-only : une clé taille+mtime sur les aqm_eval.csv suffit à invalider.
// Avant ce module, /api/integrity re-hachait tout le corpus À CHAQUE APPEL
// (deux fois : ProvenanceSHA puis ProvenanceHash8) — ~430–1100 ms par requête,
// payés sur le premier rendu du cockpit. /api/results était mémoïsé mais
// froid au premier appel (~3,3 s). warmCaches() paie ce coût au démarrage.
type corpusCache struct {
	sync.Mutex
	key       string
	integrity map[string]any
}

// corpusStamp — clé peu coûteuse : la même empreinte que l'ETag /api/results.
func corpusStamp() string {
	return results.Fingerprint(results.RunsRoot(), "")
}

// integrityPayload — construit le même document que l'ancien handler, mais
// une seule passe de hachage : hash8 dérive du SHA complet, pas d'un second
// parcours du corpus. Mémoïsé par empreinte du corpus.
func (s *Server) integrityPayload() map[string]any {
	stamp := corpusStamp()
	s.corpus.Lock()
	if stamp != "" && stamp == s.corpus.key && s.corpus.integrity != nil {
		v := s.corpus.integrity
		s.corpus.Unlock()
		return v
	}
	s.corpus.Unlock()

	v := buildIntegrity()
	if stamp != "" {
		s.corpus.Lock()
		s.corpus.key, s.corpus.integrity = stamp, v
		s.corpus.Unlock()
	}
	return v
}

// buildIntegrity — le calcul réel, sans cache (tests et premier appel).
func buildIntegrity() map[string]any {
	runs, _ := filepath.Glob(filepath.Join(results.RunsRoot(), "*"))
	if len(runs) == 0 {
		return map[string]any{"available": false, "reason": "intégrité disponible après gel (jalon M2)"}
	}
	sort.Strings(runs)
	valid, degraded, quarantined, manifests := 0, 0, 0, 0
	runIDs := []string{}
	type runBreak struct {
		Run         string `json:"run"`
		Rows        int    `json:"rows"`
		Valid       int    `json:"valid"`
		Degraded    int    `json:"degraded"`
		Quarantined int    `json:"quarantined"`
	}
	breakdown := []runBreak{}
	for _, p := range runs {
		if st, err := os.Stat(filepath.Join(p, "manifest.json")); err == nil && !st.IsDir() {
			manifests++
		}
		gs, _ := results.Scan(results.RunsRoot(), filepath.Base(p))
		b := runBreak{Run: filepath.Base(p)}
		for _, g := range gs {
			strict := g.Count - g.Quarantined - g.Degraded
			valid += strict
			degraded += g.Degraded
			quarantined += g.Quarantined
			b.Rows += g.Count
			b.Valid += strict
			b.Degraded += g.Degraded
			b.Quarantined += g.Quarantined
		}
		breakdown = append(breakdown, b)
		runIDs = append(runIDs, filepath.Base(p))
	}
	// run-<unix> : l'ordre lexical inverse est l'ordre chronologique inverse.
	reverse(breakdown)
	reverse(runIDs)
	updated := ""
	if files, _ := filepath.Glob(filepath.Join(results.RunsRoot(), "*", "aqm_eval.csv")); len(files) > 0 {
		sort.Strings(files)
		if st, err := os.Stat(files[len(files)-1]); err == nil {
			updated = st.ModTime().UTC().Format(time.RFC3339)
		}
	}
	sha := figures.ProvenanceSHA(results.RunsRoot())
	hash8 := ""
	if len(sha) >= 8 {
		hash8 = sha[:8]
	}
	// Vérification des manifestes : lecture seule, aucun fichier du corpus gelé
	// n'est modifié. Un manifeste absent ou un hash divergant dégrade l'état.
	ok, statuses := verifyRunManifests()
	integrityStatus := "verified"
	if !ok {
		integrityStatus = "degraded"
	}
	return map[string]any{
		"available": true, "runs": len(runs), "manifests": manifests,
		"valid": valid, "degraded": degraded, "quarantined": quarantined,
		"run_ids": runIDs, "breakdown": breakdown, "updated": updated,
		"sha256": sha, "hash8": hash8,
		"ok": ok, "integrity": integrityStatus, "manifest_statuses": statuses,
	}
}

func reverse[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// manifestFile — une entrée de manifest.json, forme réelle du corpus gelé.
type manifestFile struct {
	File   string `json:"File"`
	SHA256 string `json:"Sha256"`
}

type runManifest struct {
	Files []manifestFile `json:"files"`
}

// verifyRunManifests — lecture seule : chaque fichier déclaré doit exister et
// son SHA-256 doit correspondre. Un run sans manifest est signalé, pas ignoré.
func verifyRunManifests() (bool, []map[string]any) {
	runs, _ := filepath.Glob(filepath.Join(results.RunsRoot(), "*"))
	sort.Strings(runs)
	statuses := []map[string]any{}
	ok := true
	for _, p := range runs {
		st := map[string]any{"run": filepath.Base(p), "status": "verified"}
		data, err := os.ReadFile(filepath.Join(p, "manifest.json"))
		if err != nil {
			st["status"] = "no_manifest"
			ok = false
			statuses = append(statuses, st)
			continue
		}
		var m runManifest
		if err := json.Unmarshal(data, &m); err != nil {
			st["status"] = "invalid_manifest"
			ok = false
			statuses = append(statuses, st)
			continue
		}
		for _, f := range m.Files {
			body, err := os.ReadFile(filepath.Join(p, f.File))
			if err != nil {
				st["status"] = "missing_file"
				ok = false
				break
			}
			h := sha256.Sum256(body)
			if hex.EncodeToString(h[:]) != f.SHA256 {
				st["status"] = "hash_mismatch"
				ok = false
				break
			}
		}
		statuses = append(statuses, st)
	}
	return ok, statuses
}

// warmCaches — paie une fois le coût du corpus au démarrage du serveur réel,
// avant le premier rendu. Les tests passent par NewServer et ne paient rien.
func (s *Server) warmCaches() {
	_ = s.integrityPayload()
	_, _ = results.Scan(results.RunsRoot(), "")
}
