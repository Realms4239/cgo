// Meteolink (binaire cgo) — purge de rétention des sessions d'audit.
//
// `cgo audit purge` élague les artefacts d'audit en fin de vie. Quand un
// doute existe, on GARDE : rien d'autre qu'une session scellée reconnue
// (dossier « audit-* » contenant audit_report.json) n'est supprimable,
// jamais data/runs ni le corpus gelé, jamais un dossier scellé parmi les
// --keep plus récentes.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Realms4239/cgo/audit"
)

// purgeAuditSession — une session éligible, âge résolu et emplacement absolu.
type purgeAuditSession struct {
	name string
	path string
	when time.Time
}

// purgeAuditsRoot — DIR peut être la racine des données (celle qui contient
// audits/, défaut = CGO_DATA_DIR ou « data/ ») ou directement le répertoire
// audits. Tout le reste est refusé (exit 2 côté appelant) : on ne devine
// jamais une cible de suppression.
func purgeAuditsRoot(dir string) (string, error) {
	st, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !st.IsDir() {
		return "", fmt.Errorf("%s n'est pas un répertoire", dir)
	}
	if sub := filepath.Join(dir, "audits"); isDir(sub) {
		return sub, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "audit-") {
			return dir, nil
		}
	}
	if len(entries) == 0 {
		// Répertoire vide : racine de données sûrement fraîche — rien à faire.
		return dir, nil
	}
	return "", fmt.Errorf("%q ne contient ni audits/ ni sessions audit-* — cible refusée", dir)
}

func isDir(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

// purgeSessionTime — l'âge vit dans le nom (audit-<unix>, posé par le CLI).
// Sinon le mtime fait foi ; illisible ou aberrant → session écartée (gardée).
func purgeSessionTime(name string, info os.FileInfo) (time.Time, bool) {
	if ts, err := strconv.ParseInt(strings.TrimPrefix(name, "audit-"), 10, 64); err == nil && ts > 0 {
		return time.Unix(ts, 0).UTC(), true
	}
	if info != nil {
		return info.ModTime().UTC(), true
	}
	return time.Time{}, false
}

// listPurgeCandidates — les sessions scellées, triées de la plus récente à la
// plus ancienne. Tout ce qui n'est pas sûrement une session scellée est exclu
// ici (conservé, jamais listé supprimable).
func listPurgeCandidates(auditsRoot string) ([]purgeAuditSession, []string, error) {
	entries, err := os.ReadDir(auditsRoot)
	if err != nil {
		return nil, nil, err
	}
	kept := []string{}
	candidates := []purgeAuditSession{}
	for _, e := range entries {
		if !e.IsDir() { // les symlinks vers un dossier ont IsDir()=false : jamais suivis
			kept = append(kept, e.Name())
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "audit-") {
			kept = append(kept, name) // corpus gelé, runs, autre : intouchable
			continue
		}
		dir := filepath.Join(auditsRoot, name)
		if fi, err := os.Stat(filepath.Join(dir, audit.ReportArtifactName)); err != nil || !fi.Mode().IsRegular() {
			// Sans sceau : impossible de certifier que c'est une session à nous — on garde.
			kept = append(kept, name)
			continue
		}
		fi, err := os.Stat(dir)
		if err != nil {
			kept = append(kept, name)
			continue
		}
		when, ok := purgeSessionTime(name, fi)
		if !ok {
			kept = append(kept, name)
			continue
		}
		full := filepath.Join(auditsRoot, name)
		if filepath.Dir(full) != filepath.Clean(auditsRoot) {
			kept = append(kept, name) // ceinture : jamais hors de la racine
			continue
		}
		candidates = append(candidates, purgeAuditSession{name: name, path: full, when: when})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].when.Equal(candidates[j].when) {
			return candidates[i].name > candidates[j].name
		}
		return candidates[i].when.After(candidates[j].when)
	})
	return candidates, kept, nil
}

// purgeCmd — `cgo audit purge [--keep N] [--before YYYY-MM-DD] [--dry-run] [DIR]`
//
// Sémantique (conservatrice par défaut) :
//   - --keep N (défaut 10) : fenêtre de rétention — les N sessions scellées
//     les plus récentes sont TOUJOURS gardées, quel que soit --before.
//   - --before YYYY-MM-DD : seules les sessions strictement antérieures à
//     cette date sont supprimables ; sans lui, RIEN n'est supprimé.
//   - --dry-run : liste les suppressions proposées, ne touche à rien.
//
// Codes de sortie : 0 = ok (y compris « rien à faire »), 1 = échec de
// suppression, 2 = usage (drapeau invalide, date illisible, cible refusée).
func purgeCmd(args []string) int {
	fs := flag.NewFlagSet("audit purge", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	keep := fs.Int("keep", 10, "sessions scellées les plus récentes toujours conservées")
	before := fs.String("before", "", "supprime les sessions antérieures à cette date (YYYY-MM-DD)")
	dryRun := fs.Bool("dry-run", false, "affiche ce qui serait supprimé, ne supprime rien")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 1 {
		fmt.Fprintln(os.Stderr, "purge : trop d'arguments — cgo audit purge [--keep N] [--before YYYY-MM-DD] [--dry-run] [DIR]")
		return 2
	}
	if *keep < 0 {
		fmt.Fprintln(os.Stderr, "purge : --keep doit être ≥ 0")
		return 2
	}
	var cutoff time.Time
	if *before != "" {
		t, err := time.Parse("2006-01-02", *before)
		if err != nil {
			fmt.Fprintf(os.Stderr, "purge : --before %q illisible (attendu YYYY-MM-DD)\n", *before)
			return 2
		}
		cutoff = t
	}
	dir := audit.DataDir()
	if fs.NArg() == 1 {
		dir = fs.Arg(0)
	}
	auditsRoot, err := purgeAuditsRoot(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "purge :", err)
		return 2
	}
	candidates, kept, err := listPurgeCandidates(auditsRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, "purge :", err)
		return 1
	}

	toDelete := []purgeAuditSession{}
	for i, c := range candidates {
		if i < *keep {
			continue // fenêtre de rétention : toujours conservée
		}
		if !cutoff.IsZero() && c.when.Before(cutoff) {
			toDelete = append(toDelete, c)
		}
	}

	if len(toDelete) == 0 {
		switch {
		case cutoff.IsZero():
			fmt.Printf("purge — rien à faire (%d session(s) scellée(s) sous %s ; fenêtre --keep %d ; aucune suppression sans --before ; %d écartée(s))\n",
				len(candidates), auditsRoot, *keep, len(kept))
		default:
			fmt.Printf("purge — rien à faire (%d session(s) scellée(s) ; fenêtre --keep %d ; rien d'antérieur à %s hors fenêtre ; %d écartée(s))\n",
				len(candidates), *keep, cutoff.Format("2006-01-02"), len(kept))
		}
		return 0
	}

	if *dryRun {
		fmt.Printf("purge (à sec) — %d session(s) seraient supprimée(s), %d conservée(s) (fenêtre --keep %d), %d écartée(s) :\n",
			len(toDelete), len(candidates)-len(toDelete), *keep, len(kept))
		for _, c := range toDelete {
			fmt.Printf("  à supprimer : %s (%s)\n", c.path, c.when.Format("2006-01-02"))
		}
		fmt.Println("purge (à sec) — rien n'a été supprimé")
		return 0
	}

	failed := 0
	for _, c := range toDelete {
		if err := os.RemoveAll(c.path); err != nil {
			fmt.Fprintf(os.Stderr, "purge : suppression de %s échouée : %v\n", c.path, err)
			failed++
			continue
		}
		fmt.Printf("purge — supprimée : %s (%s)\n", c.path, c.when.Format("2006-01-02"))
	}
	fmt.Printf("purge — %d session(s) supprimée(s), %d conservée(s) (fenêtre --keep %d)\n",
		len(toDelete)-failed, len(candidates)-len(toDelete)+failed, *keep)
	if failed > 0 {
		return 1
	}
	return 0
}
