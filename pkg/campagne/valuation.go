package campagne

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Realms4239/cgo/pkg/metrics"
	"github.com/Realms4239/cgo/pkg/model"
)

// ValuationLabels — la provenance du run, gelée dans le sidecar : le
// régime (flux), le canal (clean|bursty) et le banc (legacy|dualns).
type ValuationLabels struct {
	Flows        int    `json:"flows"`
	ChannelModel string `json:"channel_model"`
	Bench        string `json:"bench"`
	ChargeSec    int    `json:"charge_sec"`
}

func (l ValuationLabels) set() bool { return l.Flows > 0 && l.ChargeSec > 0 }

// EventValuation — la valorisation d'UN événement : ratio (proxy documenté
// drops×1448 / octets, base "estimated" — jamais présenté comme mesuré) et
// les cinq scénarios contractuels. Sans primaire : une campagne ne déclare
// pas de type de lien.
type EventValuation struct {
	Cell      string                      `json:"cell"` // profile/qdisc/cc
	Bytes     uint64                      `json:"bytes"`
	Drops     uint64                      `json:"drops"`
	Ratio     metrics.RetransRatio        `json:"ratio"`
	Primary   *metrics.EconomicsScenario  `json:"primary,omitempty"`
	Scenarios []metrics.EconomicsScenario `json:"scenarios"`
}

// ValuationSidecar — valuation.json du run : étiquettes de provenance,
// entrée par événement gelé, et le disclaimer qui garde le tout honnête
// (scénarios contractuels, pas des mesures).
type ValuationSidecar struct {
	Labels     ValuationLabels  `json:"labels"`
	Note       string           `json:"note"`
	Disclaimer string           `json:"disclaimer"`
	Entries    []EventValuation `json:"entries"`
}

// writeValuationEntry — ajoute/remplace l'entrée d'une cellule dans le
// sidecar du run (idempotent par cellule : une reprise ne duplique pas).
// Le fichier vit À CÔTÉ de aqm_eval.csv : le CSV gelé garde son schéma
// immuable, l'économie entreprise voyage en annexe.
func writeValuationEntry(dir string, ev model.Event, labels ValuationLabels) error {
	v := ValuationSidecar{
		Labels: labels,
		Note:   "Campagne sans type de lien déclaré — aucun primaire marqué : les cinq scénarios se lisent côte à côte.",
	}
	if path := filepath.Join(dir, "valuation.json"); func() bool {
		raw, err := os.ReadFile(path)
		return err == nil && json.Unmarshal(raw, &v) == nil
	}() {
		// fichier existant relu : on remplace l'entrée de cette cellule
	} else {
		v.Disclaimer = metrics.BuildEconomics("", metrics.RetransRatio{}, 0, 0, 0).Disclaimer
	}
	if v.Disclaimer == "" {
		v.Disclaimer = metrics.BuildEconomics("", metrics.RetransRatio{}, 0, 0, 0).Disclaimer
	}
	entry := EventValuation{
		Cell:  ev.Profile + "/" + string(ev.Qdisc) + "/" + string(ev.CC),
		Bytes: ev.Bytes, Drops: ev.Drops,
	}
	if ev.Bytes > 0 && ev.Drops > 0 {
		wasted := float64(ev.Drops) * 1448
		entry.Ratio = metrics.RetransRatio{Ratio: wasted / float64(ev.Bytes), Basis: "estimated"}
		econ := metrics.BuildEconomics("", entry.Ratio, float64(ev.Bytes), wasted, float64(labels.ChargeSec))
		entry.Scenarios = econ.Scenarios
	}
	replaced := false
	for i := range v.Entries {
		if v.Entries[i].Cell == entry.Cell {
			v.Entries[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		v.Entries = append(v.Entries, entry)
	}
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "valuation.json"), raw, 0644)
}

// WriteValuation — point d'entrée public (tests, outillage) : écrit
// l'entrée d'un événement dans le sidecar du répertoire de run.
func WriteValuation(dir string, ev model.Event, labels ValuationLabels) error {
	if !labels.set() {
		labels.ChargeSec = model.ChargeSec
		labels.Flows = 4
		if labels.ChannelModel == "" {
			labels.ChannelModel = "clean"
		}
		if labels.Bench == "" {
			labels.Bench = "legacy"
		}
	}
	return writeValuationEntry(dir, ev, labels)
}

// labelsFromDeps — les étiquettes du chemin de run : mêmes défauts que
// buildMatrixConfig, pour que le sidecar et le manifeste ne se contredisent
// jamais sur la provenance d'un même run.
func labelsFromDeps(deps Deps) ValuationLabels {
	labels := ValuationLabels{
		Flows: deps.Flows, ChannelModel: deps.ChannelModel, Bench: deps.Bench,
		ChargeSec: deps.ChargeSec,
	}
	if labels.Flows <= 0 {
		labels.Flows = 4
	}
	if labels.ChannelModel == "" {
		labels.ChannelModel = "clean"
	}
	if labels.Bench == "" {
		labels.Bench = "legacy"
	}
	if labels.ChargeSec == 0 {
		labels.ChargeSec = model.ChargeSec
	}
	return labels
}
