package campagne

import (
	"encoding/json"
	"fmt"
	"os"
)

// defaultTopologyPath — chemin écrit par kit/run-dualns.sh après un
// `up`/`serve` réussi, surchargé par CGO_TOPOLOGY_JSON.
const defaultTopologyPath = "/tmp/cgo-dualns-topology.json"

// benchTopology est la portion du fichier topologie qui nous concerne :
// le mode de façonnage que le banc a réellement posé (IFB ingress ou egress).
type benchTopology struct {
	ShapeMode string `json:"shape_mode"`
}

// CheckBenchTopology refuse une cellule dont le mode de façonnage contredit
// la topologie que le banc a réellement construite : banc ingress + campagne
// egress (ou l'inverse) fait mesurer un chemin non façonné en le croyant
// façonné — fraude de mesure, refusée par construction. expected vide ⇒
// egress (défaut de la couche Go). Un fichier ABSENT (banc plus ancien) reste
// légal : nil. Un fichier présent mais illisible, invalide ou sans
// shape_mode est un échec fermé.
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
			return nil // pas de fichier topologie : rien à contredire
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
	return nil
}
