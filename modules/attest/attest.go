// Package attest — vérification de scellés SHA-256 (provenance additive).
// Le corpus gelé n'est jamais modifié ; on vérifie seulement.
package attest

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// Entry — un scellé : chemin de fichier et son sha256 attendu (hex).
type Entry struct {
	Path   string
	SHA256 string
}

// Parse lit un manifeste : une ligne par fichier, format
// "<chemin> <sha256hex>" (espaces multiples tolérés, # = commentaire).
func Parse(r *os.File) ([]Entry, error) {
	var entries []Entry
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Fields(line)
		if len(f) != 2 {
			return nil, fmt.Errorf("ligne malformée (attendu \"<chemin> <sha256>\") : %q", line)
		}
		entries = append(entries, Entry{Path: f[0], SHA256: strings.ToLower(f[1])})
	}
	return entries, sc.Err()
}

// VerifyManifest vérifie chaque fichier du manifeste contre son scellé.
// Premier écart = erreur décrivant le fichier ; nil si tout passe.
func VerifyManifest(manifestPath string) error {
	f, err := os.Open(manifestPath)
	if err != nil {
		return fmt.Errorf("ouvre le manifeste : %w", err)
	}
	defer f.Close()
	entries, err := Parse(f)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := VerifyFile(e); err != nil {
			return err
		}
	}
	return nil
}

// VerifyFile compare le sha256 réel d'un fichier à son scellé.
// Fichier absent ou différent = erreur (jamais de zéro inventé).
func VerifyFile(e Entry) error {
	data, err := os.ReadFile(e.Path)
	if err != nil {
		return fmt.Errorf("%s : %w", e.Path, err)
	}
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if got != e.SHA256 {
		return fmt.Errorf("%s : scellé %q, réel %q", e.Path, e.SHA256, got)
	}
	return nil
}
