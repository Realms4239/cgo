package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/Realms4239/cgo/audit"
)

// FlowRow — une ligne de flows.csv scellé. srtt_us est en microsecondes ;
// une valeur absente ou invalide reste une absence (NaN), jamais un zéro mesuré.
type FlowRow struct {
	SRTTus float64
}

// upperMs — bornes supérieures half-open des profils. P3 n'a pas de plafond.
var upperMs = map[string]float64{"P1": 100, "P2": 200, "P3": math.Inf(1), "P4": 60}

// BucketFlows — range chaque srtt valide dans les bandes [floor, upper).
// Une bande sans échantillon reste explicitement vide : aucune synthèse.
func BucketFlows(rows []FlowRow, floors map[string]float64) map[string][]float64 {
	buckets := make(map[string][]float64, len(floors))
	for p := range floors {
		buckets[p] = nil
	}
	for _, row := range rows {
		if math.IsNaN(row.SRTTus) || math.IsInf(row.SRTTus, 0) || row.SRTTus <= 0 {
			continue
		}
		ms := row.SRTTus / 1000
		for p, floor := range floors {
			upper, ok := upperMs[p]
			if !ok {
				upper = math.Inf(1)
			}
			// Ruling 2026-11-23 : P4 a priorité sur [40,60), P1 l'exclut.
			if p == "P1" && ms >= 40 && ms < 60 {
				continue
			}
			if ms >= floor && ms < upper {
				buckets[p] = append(buckets[p], ms)
			}
		}
	}
	return buckets
}

// ValidateBand — refuse une bande sous le seuil, en nommant le profil et le
// nombre observé. Le profil est optionnel pour garder l'appel à deux arguments.
func ValidateBand(samples []float64, min int, profile ...string) error {
	if len(samples) >= min {
		return nil
	}
	name := "band"
	if len(profile) > 0 && profile[0] != "" {
		name = profile[0]
	}
	return fmt.Errorf("%s has %d samples (< %d)", name, len(samples), min)
}

var telemetryHeader = []string{"phase", "offset_ms", "rtt_ms", "small_ms", "bulk_bytes"}

// WriteTelemetry — écrit <dir>/audits/<traceID>-<profile>/telemetry.csv dans le
// format exact lu par `cgo calibrate`, et rend le sha256 des octets écrits.
func WriteTelemetry(dir, traceID, profile string, samples []float64) (string, error) {
	id, err := audit.ArtifactID(traceID + "-" + profile)
	if err != nil {
		return "", err
	}
	path, err := audit.ArtifactFile(dir, id, audit.TelemetryArtifactName)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(telemetryHeader); err != nil {
		return "", err
	}
	for i, s := range samples {
		if math.IsNaN(s) || math.IsInf(s, 0) || s < 0 {
			return "", fmt.Errorf("tracemodel: échantillon %d invalide %v", i, s)
		}
		// small_ms et bulk_bytes vides : aucune mesure fabriquée pour un capteur absent.
		if err := w.Write([]string{"baseline", fmt.Sprintf("%d", i*5000), fmt.Sprintf("%.3f", s), "", ""}); err != nil {
			return "", err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	sum := sha256.Sum256(buf.Bytes())

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	if _, err := f.Write(buf.Bytes()); err != nil {
		_ = f.Close()
		return "", err
	}
	// Le sceau lit ce fichier juste après : fsync avant le retour.
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return hex.EncodeToString(sum[:]), nil
}
