// emdcheck calcule une distance de Wasserstein réelle entre deux jeux scellés :
// rtt_ms du terrain (ms) et srtt_us du banc (µs converti en ms). Le calcul, la
// normalisation et le refus viennent de metrics.DistributionEMD ; aucun seuil
// n'est réécrit ici.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/Realms4239/cgo/metrics"
)

type options struct {
	field string
	bench string
	out   string
	minUs int64
}

// Report — le nombre mesuré, son plafond déclaré et la provenance des deux
// fichiers. Le state applique le plafond déclaré, l'effectif minimal et le refus
// de la bibliothèque ; il ne verdit jamais un dépassement.
type Report struct {
	FieldSamples int     `json:"field_samples"`
	BenchSamples int     `json:"bench_samples"`
	EMD          float64 `json:"emd"`
	EMDCeiling   float64 `json:"emd_ceiling"`
	State        string  `json:"state"`
	Note         string  `json:"note"`
	FieldSHA256  string  `json:"field_sha256"`
	BenchSHA256  string  `json:"bench_sha256"`
}

func parseOptions(args []string) (options, error) {
	var o options
	fs := flag.NewFlagSet("emdcheck", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&o.field, "field", "", "field telemetry CSV (rtt_ms)")
	fs.StringVar(&o.bench, "bench", "", "bench observatory flows CSV (srtt_us)")
	fs.Int64Var(&o.minUs, "min-us", 600000, "minimum srtt_us of the profile band")
	fs.StringVar(&o.out, "out", "", "JSON report to write")
	if err := fs.Parse(args); err != nil {
		return options{}, err
	}
	if o.field == "" || o.bench == "" || o.out == "" {
		return options{}, fmt.Errorf("--field, --bench and --out are required")
	}
	return o, nil
}

func sha256File(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// readField — colonne rtt_ms en ms ; en-tête sauté, lignes vides ou non
// numériques ignorées (jamais transformées en 0 ms mesuré).
func readField(path string) ([]float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1 // les lignes incomplètes sont ignorées, pas fatales
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("emdcheck: en-tête %s: %w", path, err)
	}
	idx := slices.Index(header, "rtt_ms")
	if idx < 0 {
		return nil, fmt.Errorf("emdcheck: %s: colonne rtt_ms absente", path)
	}
	var out []float64
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("emdcheck: %s: %w", path, err)
		}
		if idx >= len(rec) {
			continue
		}
		raw := strings.TrimSpace(rec[idx])
		if raw == "" {
			continue
		}
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			continue
		}
		out = append(out, v)
	}
	return out, nil
}

// readBench — lignes F seulement, colonne srtt_us en µs convertie en ms ;
// srtt_us < minUs est hors bande et ignoré.
func readBench(path string, minUs int64) ([]float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("emdcheck: en-tête %s: %w", path, err)
	}
	kindIdx := slices.Index(header, "kind")
	srttIdx := slices.Index(header, "srtt_us")
	if kindIdx < 0 || srttIdx < 0 {
		return nil, fmt.Errorf("emdcheck: %s: colonnes kind/srtt_us absentes", path)
	}
	var out []float64
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("emdcheck: %s: %w", path, err)
		}
		if kindIdx >= len(rec) || srttIdx >= len(rec) || strings.TrimSpace(rec[kindIdx]) != "F" {
			continue
		}
		raw := strings.TrimSpace(rec[srttIdx])
		if raw == "" {
			continue
		}
		us, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || us < minUs {
			continue
		}
		out = append(out, float64(us)/1000)
	}
	return out, nil
}

func buildReport(fieldPath, benchPath string, minUs int64) (Report, error) {
	field, err := readField(fieldPath)
	if err != nil {
		return Report{}, err
	}
	bench, err := readBench(benchPath, minUs)
	if err != nil {
		return Report{}, err
	}
	fieldSHA, err := sha256File(fieldPath)
	if err != nil {
		return Report{}, err
	}
	benchSHA, err := sha256File(benchPath)
	if err != nil {
		return Report{}, err
	}

	// Un seul appel : normalisation et refus appartiennent à la bibliothèque.
	emd, emdErr := metrics.DistributionEMD(field, bench)
	rep := Report{
		FieldSamples: len(field),
		BenchSamples: len(bench),
		EMD:          emd,
		EMDCeiling:   metrics.FidelityEMDMax,
		FieldSHA256:  fieldSHA,
		BenchSHA256:  benchSHA,
	}
	switch {
	case emdErr != nil:
		rep.State = metrics.FidelityDiscard
		rep.Note = fmt.Sprintf("EMD refused by metrics.DistributionEMD: %v; no value trusted", emdErr)
	case emd > rep.EMDCeiling:
		// Preuve contraire : un dépassement n'est jamais masqué par un
		// effectif insuffisant (même sémantique que DelayAccuracy).
		rep.State = metrics.FidelityDiscard
		rep.Note = fmt.Sprintf("measured EMD %.6f exceeds declared ceiling %.6f (metrics.FidelityEMDMax): distributions differ beyond the declared bound", emd, rep.EMDCeiling)
	case len(field) < metrics.FidelityMinSamples || len(bench) < metrics.FidelityMinSamples:
		rep.State = metrics.FidelityBounded
		rep.Note = fmt.Sprintf("below FidelityMinSamples=%d (field=%d, bench=%d): EMD %.6f within ceiling %.6f but not trusted",
			metrics.FidelityMinSamples, len(field), len(bench), emd, rep.EMDCeiling)
	default:
		rep.State = metrics.FidelityTrust
		rep.Note = fmt.Sprintf("field=%d, bench=%d >= FidelityMinSamples=%d; EMD %.6f within declared ceiling %.6f (metrics.DistributionEMD)",
			len(field), len(bench), metrics.FidelityMinSamples, emd, rep.EMDCeiling)
	}
	return rep, nil
}

func main() {
	o, err := parseOptions(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "emdcheck:", err)
		os.Exit(2)
	}
	rep, err := buildReport(o.field, o.bench, o.minUs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "emdcheck:", err)
		os.Exit(1)
	}
	body, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "emdcheck:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(o.out, append(body, '\n'), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "emdcheck:", err)
		os.Exit(1)
	}
}
