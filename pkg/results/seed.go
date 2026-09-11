package results

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//go:embed runs-frozen.tar.gz
var runsFrozen []byte

// SeedRuns garantit un corpus de runs gelés sous <root>/data/runs.
// Si au moins un aqm_eval.csv existe déjà, ne fait rien (seeded=false) :
// les runs vivants et l'historique local priment toujours sur le gel.
// Sinon extrait l'archive intégrée au binaire (portabilité : un binaire
// seul suffit sur une machine vierge). Seules les entrées régulières sous
// data/runs/ sont acceptées (anti-traversal strict).
func SeedRuns(root string) (seeded bool, n int, err error) {
	runsDir := filepath.Join(root, "data", "runs")
	if files, _ := filepath.Glob(filepath.Join(runsDir, "*", "aqm_eval.csv")); len(files) > 0 {
		return false, 0, nil
	}
	zr, err := gzip.NewReader(bytes.NewReader(runsFrozen))
	if err != nil {
		return false, 0, fmt.Errorf("seed: %w", err)
	}
	defer zr.Close()
	tr := tar.NewReader(zr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, n, fmt.Errorf("seed: %w", err)
		}
		name := filepath.ToSlash(hdr.Name)
		if !strings.HasPrefix(name, "data/runs/") || hdr.Typeflag != tar.TypeReg {
			continue
		}
		rel := strings.TrimPrefix(name, "data/runs/")
		if rel == "" || strings.Contains(rel, "..") || filepath.IsAbs(rel) {
			continue
		}
		dst := filepath.Join(runsDir, filepath.FromSlash(rel))
		if !strings.HasPrefix(dst, runsDir+string(os.PathSeparator)) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return false, n, fmt.Errorf("seed: %w", err)
		}
		f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			if os.IsExist(err) {
				continue
			}
			return false, n, fmt.Errorf("seed: %w", err)
		}
		_, werr := io.Copy(f, tr)
		cerr := f.Close()
		if werr != nil {
			return false, n, fmt.Errorf("seed: %w", werr)
		}
		if cerr != nil {
			return false, n, fmt.Errorf("seed: %w", cerr)
		}
		n++
	}
	if n == 0 {
		return false, 0, fmt.Errorf("seed: archive intégrée vide ou illisible")
	}
	return true, n, nil
}
