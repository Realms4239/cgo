package campagne

import (
	"encoding/json"
	"fmt"
	"os"
)

// defaultTopologyPath — chemin écrit par kit/run-dualns.sh après un
// `up`/`serve` réussi, surchargé par CGO_TOPOLOGY_JSON.
const defaultTopologyPath = "/tmp/cgo-dualns-topology.json"

// benchTopology est le contrat écrit par le banc après un `up`/`serve` réussi :
// façonnage réellement posé (IFB ingress ou egress) et physique déclarée.
// CheckBenchTopology vérifie le contrat complet ; champ manquant ou incohérent
// ⇒ échec fermé.
type benchTopology struct {
	ShapeMode string  `json:"shape_mode"`
	RTTMode   string  `json:"rtt_mode"`
	DelayMs   float64 `json:"delay_ms"`
	RateMbps  float64 `json:"rate_mbps"`
	IFBActive bool    `json:"ifb_active"`
}

// CheckBenchTopology refuse une cellule dont la topologie contredit le contrat
// du banc : façonnage (IFB ingress ou egress), convention RTT (per_hop_half),
// délai et débit positifs, cohérence ifb_active/direction. Banc ingress +
// campagne egress (ou l'inverse) fait mesurer un chemin non façonné en le
// croyant façonné — fraude de mesure, refusée par construction. expected vide ⇒
// egress (défaut de la couche Go). Un fichier ABSENT est désormais un échec
// fermé : on ne peut certifier la topologie d'un banc plus ancien. Tout fichier
// présent mais illisible, invalide ou incomplet est également refusé.
func CheckBenchTopology(expected string) error {
	if expected == "" {
		expected = "egress"
	}
	path := os.Getenv("CGO_TOPOLOGY_JSON")
	if path == "" {
		path = defaultTopologyPath
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("topologie banc absente (%s): fail closed", path)
		}
		return fmt.Errorf("topologie banc illisible (%s): %w", path, err)
	}
	var topo benchTopology
	if err := json.Unmarshal(data, &topo); err != nil {
		return fmt.Errorf("topologie banc invalide (%s): %w", path, err)
	}
	switch topo.ShapeMode {
	case "ingress", "egress":
	default:
		return fmt.Errorf("topologie banc (%s): shape_mode %q absent ou invalide (ingress|egress requis)", path, topo.ShapeMode)
	}
	if topo.ShapeMode != expected {
		return fmt.Errorf("façonnage contradictoire : banc %s, campagne %s — cellule refusée (la campagne mesurerait un chemin non façonné en le croyant façonné)", topo.ShapeMode, expected)
	}
	// Le contrat physique est vérifié après l'accord de façonnage, pour
	// conserver le message de contradiction explicite ci-dessus.
	if topo.RTTMode != "per_hop_half" {
		return fmt.Errorf("topologie banc (%s): rtt_mode %q invalide (per_hop_half requis)", path, topo.RTTMode)
	}
	if !(topo.DelayMs > 0) || !(topo.RateMbps > 0) {
		return fmt.Errorf("topologie banc (%s): delay_ms et rate_mbps doivent être > 0", path)
	}
	if topo.IFBActive != (expected == "ingress") {
		return fmt.Errorf("topologie banc (%s): ifb_active=%v contredit la direction %q", path, topo.IFBActive, expected)
	}
	return nil
}
