package qdisc

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Realms4239/cgo/model"
)

// Constantes du format netem (tc/q_netem.c) : la table est lue comme des
// entiers base 0, et le noyau applique mu + (sigma/8192)*t. NETEM_DIST_MAX
// borne une entrée, pas la taille de la table (16384).
const (
	distEntries    = 1024
	distScale      = 8192.0
	distEntryMax   = 32767
	distMinSamples = 30
)

// Calibration — distribution empirique d'un profil et sa provenance. La
// table n'est pas dupliquée dans le JSON : le .dist la porte, le sidecar
// la scelle par DistSHA256.
type Calibration struct {
	Schema                int    `json:"schema"`
	Profile               string `json:"profile"`
	Basis                 string `json:"basis"`
	HopModel              string `json:"hop_model"`
	SourceAudit           string `json:"source_audit"`
	SourceReportSHA256    string `json:"source_report_sha256,omitempty"`
	SourceTelemetrySHA256 string `json:"source_telemetry_sha256"`
	SourceSite            string `json:"source_site,omitempty"`
	SourceLinkType        string `json:"source_link_type,omitempty"`
	Created               string `json:"created"`

	ProfileDelayMs float64 `json:"profile_delay_ms"`
	FloorMs        float64 `json:"floor_ms"`
	HopDelayMs     float64 `json:"hop_delay_ms"`
	ScaleMs        float64 `json:"scale_ms"`
	TableSize      int     `json:"table_size"`
	DistFile       string  `json:"dist_file"`
	DistSHA256     string  `json:"dist_sha256"`

	Samples       int     `json:"samples"`
	BelowFloorPct float64 `json:"below_floor_pct"`
	ClippedPct    float64 `json:"clipped_pct"`

	SourceP50RTTms  float64 `json:"source_p50_rtt_ms"`
	SourceP95RTTms  float64 `json:"source_p95_rtt_ms"`
	SourceP999RTTms float64 `json:"source_p999_rtt_ms"`
	SourceMaxRTTms  float64 `json:"source_max_rtt_ms"`

	RoundTripP50RTTms  float64 `json:"roundtrip_p50_rtt_ms"`
	RoundTripP95RTTms  float64 `json:"roundtrip_p95_rtt_ms"`
	RoundTripP999RTTms float64 `json:"roundtrip_p999_rtt_ms"`

	Table []int16 `json:"-"`
}

// BuildCalibration — RTT baseline → table netem. Le délai du profil est un
// RTT (convention du corpus) : le résidu est RTT - plancher, et la table ne
// contient que ce résidu. Une trace dont la médiane ne dépasse pas le
// plancher ne décrit pas ce profil : refus, jamais de table dégénérée.
func BuildCalibration(prof model.Profile, auditID string, rtts []float64) (Calibration, error) {
	if strings.TrimSpace(prof.ID) == "" {
		return Calibration{}, errors.New("calibration: profil sans identifiant")
	}
	if prof.DelayMs <= 0 {
		return Calibration{}, fmt.Errorf("calibration %s: delay_ms %.2f ≤ 0", prof.ID, prof.DelayMs)
	}
	if len(rtts) < distMinSamples {
		return Calibration{}, fmt.Errorf("calibration %s: %d échantillons baseline < %d — trace trop courte", prof.ID, len(rtts), distMinSamples)
	}
	sorted := append([]float64(nil), rtts...)
	sort.Float64s(sorted)
	for _, v := range sorted {
		if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			return Calibration{}, fmt.Errorf("calibration %s: RTT invalide %v", prof.ID, v)
		}
	}
	p50 := quantile(sorted, 0.50)
	if p50 <= prof.DelayMs {
		return Calibration{}, fmt.Errorf("calibration %s: p50 RTT %.2f ms ≤ plancher %.2f ms — la trace ne correspond pas au profil", prof.ID, p50, prof.DelayMs)
	}

	residuals := make([]float64, len(sorted))
	below := 0
	for i, v := range sorted {
		r := v - prof.DelayMs
		if r < 0 {
			below++
			r = 0
		}
		residuals[i] = r
	}
	p999 := quantile(residuals, 0.999)
	if p999 <= 0 {
		return Calibration{}, fmt.Errorf("calibration %s: queue résiduelle nulle", prof.ID)
	}

	scale := p999 * distScale / distEntryMax
	table := make([]int16, distEntries)
	clipped := 0
	for i := range table {
		q := quantile(residuals, (float64(i)+0.5)/float64(distEntries))
		v := int(math.Round(q / scale * distScale))
		if v < 0 {
			v = 0
		}
		if v > distEntryMax {
			v = distEntryMax
			clipped++
		}
		table[i] = int16(v)
	}

	cal := Calibration{
		Schema:          1,
		Profile:         prof.ID,
		Basis:           "measured",
		HopModel:        "rtt_floor_split",
		SourceAudit:     auditID,
		ProfileDelayMs:  prof.DelayMs,
		FloorMs:         prof.DelayMs,
		HopDelayMs:      prof.DelayMs / 2,
		ScaleMs:         scale,
		TableSize:       distEntries,
		DistFile:        prof.ID + ".dist",
		Samples:         len(sorted),
		BelowFloorPct:   100 * float64(below) / float64(len(sorted)),
		ClippedPct:      100 * float64(clipped) / float64(distEntries),
		SourceP50RTTms:  p50,
		SourceP95RTTms:  quantile(sorted, 0.95),
		SourceP999RTTms: quantile(sorted, 0.999),
		SourceMaxRTTms:  sorted[len(sorted)-1],
		Table:           table,
	}
	cal.recomputeRoundTrip()
	return cal, nil
}

func quantile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	i := int(math.Round(p * float64(len(sorted)-1)))
	if i < 0 {
		i = 0
	}
	if i >= len(sorted) {
		i = len(sorted) - 1
	}
	return sorted[i]
}

func (c *Calibration) recomputeRoundTrip() {
	if len(c.Table) == 0 || c.ScaleMs <= 0 {
		return
	}
	vals := make([]float64, len(c.Table))
	for i, v := range c.Table {
		vals[i] = c.FloorMs + c.ScaleMs*float64(v)/distScale
	}
	sort.Float64s(vals)
	c.RoundTripP50RTTms = quantile(vals, 0.50)
	c.RoundTripP95RTTms = quantile(vals, 0.95)
	c.RoundTripP999RTTms = quantile(vals, 0.999)
}

// Render — format netem natif : un entier décimal par ligne, sans padding
// (strtol base 0 lirait « 010 » comme 8) ni commentaire.
func (c Calibration) Render() []byte {
	var b strings.Builder
	for _, v := range c.Table {
		b.WriteString(strconv.Itoa(int(v)))
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

// WriteCalibration — écrit <profil>.dist puis <profil>.json, chacun remplacé
// atomiquement. Le sidecar scelle les octets du .dist ; un .dist orphelin
// (sidecar absent) est refusé à la lecture, jamais deviné.
func WriteCalibration(dir string, cal Calibration) error {
	if err := safeCalibrationName(cal.Profile); err != nil {
		return err
	}
	if cal.Schema != 1 {
		return fmt.Errorf("calibration %s: schema %d, attendu 1", cal.Profile, cal.Schema)
	}
	if cal.Basis != "measured" {
		return fmt.Errorf("calibration %s: basis %q, attendu measured", cal.Profile, cal.Basis)
	}
	if cal.HopModel != "rtt_floor_split" {
		return fmt.Errorf("calibration %s: hop_model %q, attendu rtt_floor_split", cal.Profile, cal.HopModel)
	}
	if cal.TableSize != distEntries || len(cal.Table) != distEntries {
		return fmt.Errorf("calibration %s: table %d entrées (sidecar %d), attendu %d", cal.Profile, len(cal.Table), cal.TableSize, distEntries)
	}
	if cal.ScaleMs <= 0 || cal.FloorMs <= 0 {
		return fmt.Errorf("calibration %s: échelle/plancher invalide", cal.Profile)
	}
	if cal.DistFile == "" {
		cal.DistFile = cal.Profile + ".dist"
	}
	if cal.DistFile != cal.Profile+".dist" {
		return fmt.Errorf("calibration %s: fichier %q incohérent", cal.Profile, cal.DistFile)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("calibration: répertoire %s: %w", dir, err)
	}

	raw := cal.Render()
	sum := sha256.Sum256(raw)
	cal.DistSHA256 = hex.EncodeToString(sum[:])
	if err := writeFileAtomic(filepath.Join(dir, cal.DistFile), raw); err != nil {
		return err
	}
	sidecar, err := json.MarshalIndent(cal, "", "  ")
	if err != nil {
		return fmt.Errorf("calibration %s: sidecar: %w", cal.Profile, err)
	}
	sidecar = append(sidecar, '\n')
	return writeFileAtomic(filepath.Join(dir, cal.Profile+".json"), sidecar)
}

func writeFileAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("calibration: écriture %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("calibration: renommage %s: %w", path, err)
	}
	return nil
}

func safeCalibrationName(id string) error {
	if id == "" || id == "." || id == ".." || id != filepath.Base(id) || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("calibration: identifiant invalide %q", id)
	}
	return nil
}

// LoadCalibration — lecture fail-closed : sidecar présent, profil exact,
// .dist présent et de hash identique. Toute divergence arrête la campagne
// AVANT la première mutation qdisc ; jamais de repli synthétique.
func LoadCalibration(dir, profile string) (Calibration, error) {
	if err := safeCalibrationName(profile); err != nil {
		return Calibration{}, err
	}
	raw, err := os.ReadFile(filepath.Join(dir, profile+".json"))
	if err != nil {
		return Calibration{}, fmt.Errorf("calibration %s: sidecar absent (%w) — lancer cgo calibrate", profile, err)
	}
	var cal Calibration
	if err := json.Unmarshal(raw, &cal); err != nil {
		return Calibration{}, fmt.Errorf("calibration %s: sidecar illisible: %w", profile, err)
	}
	if cal.Schema != 1 || cal.Profile != profile {
		return Calibration{}, fmt.Errorf("calibration %s: sidecar incohérent (schema %d, profil %q)", profile, cal.Schema, cal.Profile)
	}
	if cal.Basis != "measured" {
		return Calibration{}, fmt.Errorf("calibration %s: basis %q, attendu measured", profile, cal.Basis)
	}
	if cal.HopModel != "rtt_floor_split" {
		return Calibration{}, fmt.Errorf("calibration %s: hop_model %q, attendu rtt_floor_split", profile, cal.HopModel)
	}
	if cal.FloorMs <= 0 || cal.ScaleMs <= 0 || cal.TableSize != distEntries {
		return Calibration{}, fmt.Errorf("calibration %s: plancher/échelle/table invalide", profile)
	}
	if cal.DistFile != profile+".dist" {
		return Calibration{}, fmt.Errorf("calibration %s: fichier %q incohérent", profile, cal.DistFile)
	}
	dist, err := os.ReadFile(filepath.Join(dir, cal.DistFile))
	if err != nil {
		return Calibration{}, fmt.Errorf("calibration %s: .dist absent: %w", profile, err)
	}
	sum := sha256.Sum256(dist)
	if hex.EncodeToString(sum[:]) != cal.DistSHA256 {
		return Calibration{}, fmt.Errorf("calibration %s: hash .dist divergent — artefact modifié ou incomplet", profile)
	}
	if got := bytes.Count(dist, []byte{'\n'}); got != cal.TableSize {
		return Calibration{}, fmt.Errorf("calibration %s: .dist %d lignes, sidecar %d", profile, got, cal.TableSize)
	}
	return cal, nil
}
