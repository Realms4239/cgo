package results

import (
	"os"
	"path/filepath"
	"strings"
)

// La racine du corpus — UN propriétaire, lue au moment de l'appel.
//
// Avant : « data/runs » écrit en dur dans trois modules et deux binaires,
// résolu relativement au cwd — sous systemd (WorkingDirectory=/) le
// dashboard cherchait /data/runs; l'image scratch n'avait ni WORKDIR
// ni volume. CGO_DATA_DIR l'épingle partout; le défaut historique
// (« data » relatif) reste pour les postes de travail en checkout.
//
// Tout ce qui lit ou écrit le corpus passe par ces fonctions.

// DataRoot — la racine : CGO_DATA_DIR, défaut « data ».
func DataRoot() string {
	if r := strings.TrimSpace(os.Getenv("CGO_DATA_DIR")); r != "" {
		return r
	}
	return "data"
}

// RunsRoot — le corpus gelé : <DataRoot>/runs.
func RunsRoot() string {
	return filepath.Join(DataRoot(), "runs")
}

// FiguresRoot — les figures régénérables : <DataRoot>/figures.
func FiguresRoot() string {
	return filepath.Join(DataRoot(), "figures")
}

// LinkAuditCSV — le CSV append-only des audits : <DataRoot>/link_audit.csv.
func LinkAuditCSV() string {
	return filepath.Join(DataRoot(), "link_audit.csv")
}
