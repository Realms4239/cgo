package audit

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TelemetryRow — une sonde horodatée : phase, offset depuis le début de
// l'audit, RTT ping, complétion petit objet, cumul d'octets de charge.
// C'est la matière PREMIÈRE du rapport ; le sceau SHA-256 la scelle.
// Une ligne RTT nulle ET small nulle n'existe pas : une sonde perdue est
// une absence (elle ne devient jamais un « 0 ms » mesuré).
type TelemetryRow struct {
	Phase     string  `json:"phase"`
	OffsetMs  int64   `json:"offset_ms"`
	RTTms     float64 `json:"rtt_ms"`
	SmallMs   float64 `json:"small_ms"`
	BulkBytes uint64  `json:"bulk_bytes"`
}

var telemetryHeader = []string{"phase", "offset_ms", "rtt_ms", "small_ms", "bulk_bytes"}

// ArtifactID — identifiant d'artefact sûr : nom simple, jamais . ou .., sans
// séparateur ni espace de bord. Le CLI et l'API génèrent les ids, mais un
// rapport importé, une URL ou un appel direct traverse la même porte.
func ArtifactID(id string) (string, error) {
	trimmed := strings.TrimSpace(id)
	base := filepath.Base(trimmed)
	if base == "" || base == "." || base == ".." || base == string(filepath.Separator) || base != trimmed || strings.ContainsAny(trimmed, `/\`) {
		return "", fmt.Errorf("audit: identifiant invalide %q", id)
	}
	return base, nil
}

// WriteTelemetry — data/audits/<id>/telemetry.csv. Retourne le chemin pour
// que l'appelant puisse sceller ce qu'il vient d'écrire. Un id vide ou
// traversant est refusé : jamais d'écriture hors de la zone d'audits.
func WriteTelemetry(baseDir string, res *Result) (string, error) {
	id, err := ArtifactID(res.AuditID)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(baseDir, "audits", id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "telemetry.csv")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write(telemetryHeader)
	for _, r := range res.Telemetry {
		bulk := ""
		if r.BulkBytes != 0 {
			bulk = fmt.Sprintf("%d", r.BulkBytes)
		}
		_ = w.Write([]string{
			r.Phase, fmt.Sprintf("%d", r.OffsetMs),
			fmt.Sprintf("%.2f", r.RTTms), fmt.Sprintf("%.2f", r.SmallMs),
			bulk,
		})
	}
	w.Flush()
	return path, w.Error()
}
