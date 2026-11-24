package api

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/Realms4239/cgo/audit"
	"github.com/Realms4239/cgo/exporter"
	"github.com/Realms4239/cgo/metrics"
)

// nativeGates — les six portes natives du runner cmd/gates. Le nom est une
// déclaration, jamais un état : un statut n'est publié que s'il vient d'un
// artefact réel écrit par le runner.
var nativeGates = []string{"frozen", "provenance", "tests", "exit", "hygiene", "links"}

const gateStatusArtifact = "gate_status.json"

// observabilityHelp — les six séries du design, déclarées même sans échantillon.
// Un vecteur vide émet HELP/TYPE (contrat exporter) sans valeur inventée.
var observabilityHelp = map[string]string{
	"cgo_audit_info":               "Latest sealed audit report (audit_id, source, age_bucket); always 1",
	"cgo_kernel_truth_source_info": "Available kernel truth source from the sealed report; emitted only when one exists",
	"cgo_seal_present":             "1 when the real report carries both provenance seals, 0 when it carries none",
	"cgo_emd":                      "Wasserstein EMD from the sealed fidelity verdict, labelled by its live basis",
	"cgo_emd_ceiling":              "Declared EMD ceiling from metrics.FidelityEMDMax",
	"cgo_gate_status":              "1 = pass, 0 = fail from the gate runner artifact; skipped gates are omitted",
}

// metricsHandler — seul point d'entrée du collecteur /metrics. Le collecteur
// et ses sources restent dans ce fichier, avec les séries qu'il mesure vraiment.
func (s *Server) metricsHandler() http.Handler {
	return exporter.Handler(s.metricsSamples)
}

// metricsSamples — les séries que ce serveur mesure vraiment. Aucune valeur
// inventée : une série absente n'est jamais émise à zéro (contrat exporter).
func (s *Server) metricsSamples() (map[string][]exporter.Sample, map[string]string) {
	running := "0"
	if s.deps.RunningFn != nil && s.deps.RunningFn() {
		running = "1"
	}
	ks := s.kernel
	samples := map[string][]exporter.Sample{
		"cgo_build_info": {{
			Labels: map[string]string{"version": s.deps.Version, "mode": s.deps.Mode},
			Value:  "1",
		}},
		"cgo_uptime_seconds": {{
			Value: strconv.FormatFloat(time.Since(s.started).Seconds(), 'f', 3, 64),
		}},
		"cgo_sse_subscribers": {{
			Value: strconv.Itoa(s.hub.Subscribers()),
		}},
		"cgo_campaign_running": {{
			Value: running,
		}},
		"cgo_kernel_source_info": {{
			Labels: map[string]string{"source": ks.Source, "reason": ks.Reason},
			Value:  "1",
		}},
	}
	help := map[string]string{
		"cgo_build_info":         "Build identity (version, mode); always 1",
		"cgo_uptime_seconds":     "Seconds since the API server started",
		"cgo_sse_subscribers":    "Live SSE dashboard subscribers",
		"cgo_campaign_running":   "1 while a campaign matrix runs, 0 otherwise",
		"cgo_kernel_source_info": "Kernel telemetry source chosen at startup; always 1",
	}
	for name, text := range observabilityHelp {
		if _, exists := samples[name]; !exists {
			samples[name] = []exporter.Sample{}
			help[name] = text
		}
	}
	s.addObservabilitySamples(samples)
	addGateStatusSamples(samples)
	return samples, help
}

// addObservabilitySamples — audit, sceau, fidélité et plafond, tous lus depuis
// des artefacts réels ou déclarés dans le code. Rien n'est déduit d'une note.
func (s *Server) addObservabilitySamples(samples map[string][]exporter.Sample) {
	// Le plafond est une constante déclarée, pas une observation : il reste
	// publié même sans rapport, contrairement aux valeurs mesurées.
	samples["cgo_emd_ceiling"] = append(samples["cgo_emd_ceiling"], exporter.Sample{
		Value: strconv.FormatFloat(metrics.FidelityEMDMax, 'f', -1, 64),
	})

	rep, ok := latestAuditReport()
	if !ok {
		return
	}
	// « available » est un fait scellé : un source étiquetée sur un hôte sans
	// noyau n'est pas une source de vérité, même si le probe la nomme.
	if kt := rep.KernelTelemetry; kt != nil && kt.Available && kt.Source != "" && kt.Source != "unavailable" {
		samples["cgo_kernel_truth_source_info"] = append(samples["cgo_kernel_truth_source_info"], exporter.Sample{
			Labels: map[string]string{"source": kt.Source},
			Value:  "1",
		})
	}
	id, err := audit.ArtifactID(rep.AuditID)
	if err != nil {
		return // un id traversant dans un artefact importé n'est jamais un label
	}
	source := "unspecified"
	if rep.Invocation != nil && rep.Invocation.SinkMode != "" {
		source = rep.Invocation.SinkMode
	}
	samples["cgo_audit_info"] = append(samples["cgo_audit_info"], exporter.Sample{
		Labels: map[string]string{
			"audit_id":   id,
			"source":     source,
			"age_bucket": ageBucket(rep.Timestamp, time.Now()),
		},
		Value: "1",
	})

	seal := "0"
	if rep.Provenance.SourceSHA256 != "" && rep.Provenance.TelemetrySHA256 != "" {
		seal = "1"
	}
	samples["cgo_seal_present"] = append(samples["cgo_seal_present"], exporter.Sample{
		Labels: map[string]string{"audit_id": id},
		Value:  seal,
	})

	if rep.Fidelity != nil && finite(rep.Fidelity.EMD) {
		basis := rep.Fidelity.Basis
		if basis == "" {
			basis = "unspecified"
		}
		samples["cgo_emd"] = append(samples["cgo_emd"], exporter.Sample{
			Labels: map[string]string{"live_basis": basis},
			Value:  strconv.FormatFloat(rep.Fidelity.EMD, 'f', -1, 64),
		})
	}
}

// latestAuditReport — le rapport scellé le plus réellement récent (horodatage
// scellé, pas ordre de lecture du disque). Un artefacts illisible est ignoré,
// jamais remplacé par un rapport fabriqué.
func latestAuditReport() (*audit.Report, bool) {
	files, err := filepath.Glob(filepath.Join(audit.DataDir(), "audits", "*", audit.ReportArtifactName))
	if err != nil || len(files) == 0 {
		return nil, false
	}
	sort.Strings(files)
	var best *audit.Report
	bestKey := ""
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var rep audit.Report
		if err := json.Unmarshal(b, &rep); err != nil {
			continue
		}
		if _, err := audit.ArtifactID(rep.AuditID); err != nil {
			continue
		}
		key := rep.Timestamp + "\x00" + f
		if best == nil || key > bestKey {
			best, bestKey = &rep, key
		}
	}
	return best, best != nil
}

// ageBucket — âge du rapport scellé. Une horodatage absent, invalide ou dans
// le futur est une absence nommée, jamais « fresh » par défaut.
func ageBucket(timestamp string, now time.Time) string {
	if timestamp == "" {
		return "unknown"
	}
	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return "unknown"
	}
	age := now.Sub(ts)
	if age < 0 {
		return "unknown"
	}
	if age < 24*time.Hour {
		return "fresh"
	}
	if age < 7*24*time.Hour {
		return "recent"
	}
	return "stale"
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// addGateStatusSamples — statut 0/1 uniquement depuis l'artefact écrit par le
// runner. « skip » et une porte inconnue sont des absences : ni vert, ni rouge.
func addGateStatusSamples(samples map[string][]exporter.Sample) {
	b, err := os.ReadFile(filepath.Join(audit.DataDir(), gateStatusArtifact))
	if err != nil {
		return
	}
	var raw map[string]string
	if err := json.Unmarshal(b, &raw); err != nil {
		return
	}
	for _, gate := range nativeGates {
		var value string
		switch raw[gate] {
		case "pass":
			value = "1"
		case "fail":
			value = "0"
		default:
			continue
		}
		samples["cgo_gate_status"] = append(samples["cgo_gate_status"], exporter.Sample{
			Labels: map[string]string{"gate": gate},
			Value:  value,
		})
	}
}
