// Package attest — vérification de scellés SHA-256 (provenance additive).
// Le corpus gelé n'est jamais modifié ; on vérifie seulement.
package attest

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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
// Un scellé doit être 64 caractères hexadécimaux ; un manifeste sans
// aucune entrée est une erreur (jamais de succès par défaut).
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
		if err := validateSHA256(f[1]); err != nil {
			return nil, fmt.Errorf("ligne malformée (scellé sha256 invalide) : %q", line)
		}
		entries = append(entries, Entry{Path: f[0], SHA256: strings.ToLower(f[1])})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("manifeste vide : aucune entrée à vérifier")
	}
	return entries, nil
}

// validateSHA256 exige 64 hexadécimaux (32 octets).
func validateSHA256(s string) error {
	if len(s) != 64 {
		return fmt.Errorf("scellé %d caractères, attendu 64", len(s))
	}
	if _, err := hex.DecodeString(s); err != nil {
		return fmt.Errorf("scellé non hexadécimal : %w", err)
	}
	return nil
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

// VerifyFile compare le sha256 réel d'un fichier à son scellé (flux, sans
// charger le fichier en mémoire). Fichier absent ou différent = erreur.
func VerifyFile(e Entry) error {
	if err := validateSHA256(e.SHA256); err != nil {
		return fmt.Errorf("%s : %w", e.Path, err)
	}
	f, err := os.Open(e.Path)
	if err != nil {
		return fmt.Errorf("%s : %w", e.Path, err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("%s : %w", e.Path, err)
	}
	got := hex.EncodeToString(h.Sum(nil))
	if got != strings.ToLower(e.SHA256) {
		return fmt.Errorf("%s : scellé %q, réel %q", e.Path, e.SHA256, got)
	}
	return nil
}
