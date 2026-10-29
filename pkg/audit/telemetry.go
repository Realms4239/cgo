package audit

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
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

// ReadTelemetry — lecture stricte de data/audits/<id>/telemetry.csv.
// L'en-tête doit être exactement celui du writer : une colonne déplacée
// changerait le sens des nombres. Une sonde absente reste absente (0) ;
// elle n'est jamais réinventée en mesure.
func ReadTelemetry(baseDir, id string) ([]TelemetryRow, error) {
	safe, err := ArtifactID(id)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(baseDir, "audits", safe, "telemetry.csv")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("audit: lecture télémétrie %s: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = len(telemetryHeader)
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("audit: en-tête télémétrie %s: %w", path, err)
	}
	if !slices.Equal(header, telemetryHeader) {
		return nil, fmt.Errorf("audit: en-tête télémétrie %s = %v, attendu %v", path, header, telemetryHeader)
	}

	var rows []TelemetryRow
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("audit: ligne télémétrie %s: %w", path, err)
		}
		row, err := parseTelemetryRow(rec)
		if err != nil {
			return nil, fmt.Errorf("audit: ligne télémétrie %s: %w", path, err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func parseTelemetryRow(rec []string) (TelemetryRow, error) {
	var row TelemetryRow
	row.Phase = strings.TrimSpace(rec[0])
	if row.Phase == "" {
		return row, fmt.Errorf("phase vide")
	}
	offset, err := strconv.ParseInt(strings.TrimSpace(rec[1]), 10, 64)
	if err != nil || offset < 0 {
		return row, fmt.Errorf("offset_ms invalide %q", rec[1])
	}
	row.OffsetMs = offset

	for _, f := range []struct {
		name string
		dst  *float64
		raw  string
	}{
		{"rtt_ms", &row.RTTms, rec[2]},
		{"small_ms", &row.SmallMs, rec[3]},
	} {
		s := strings.TrimSpace(f.raw)
		if s == "" {
			continue
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil || v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			return row, fmt.Errorf("%s invalide %q", f.name, f.raw)
		}
		*f.dst = v
	}

	if s := strings.TrimSpace(rec[4]); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return row, fmt.Errorf("bulk_bytes invalide %q", rec[4])
		}
		row.BulkBytes = v
	}
	if row.RTTms == 0 && row.SmallMs == 0 && row.BulkBytes == 0 {
		return row, fmt.Errorf("ligne sans mesure (rtt, small et bulk vides)")
	}
	return row, nil
}
