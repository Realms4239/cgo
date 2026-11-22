package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// Calibration — projection honnête d'un sidecar vérifié pour GET /api/calibrations.
// Aucune entrée n'est synthétisée : une ligne n'apparaît que si le hash du
// sidecar scelle les octets du .dist réellement présent.
type Calibration struct {
	Name     string  `json:"name"`
	SHA256   string  `json:"sha256"`
	Samples  int     `json:"samples"`
	P50Ms    float64 `json:"p50_ms"`
	Verified bool    `json:"verified"`
}

// calibrationSidecar — le contrat de lecture. Les noms courts (sha256, p50_ms)
// sont ceux du contrat HTTP ; les noms longs sont ceux du sidecar qdisc réel.
type calibrationSidecar struct {
	SHA256     string  `json:"sha256"`
	DistSHA256 string  `json:"dist_sha256"`
	Samples    int     `json:"samples"`
	P50Ms      float64 `json:"p50_ms"`
	SourceP50  float64 `json:"source_p50_rtt_ms"`
}

// ListCalibrations — ne renvoie que les sidecars vérifiés. Un .dist orphelin,
// un sidecar illisible ou un hash divergent est écarté : jamais de repli, jamais
// de valeur fabriquée, jamais de PASS déduit d'une absence.
func ListCalibrations(dir string) []Calibration {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []Calibration{}
	}
	out := make([]Calibration, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := e.Name()[:len(e.Name())-len(".json")]
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var side calibrationSidecar
		if err := json.Unmarshal(raw, &side); err != nil {
			continue
		}
		sha := side.SHA256
		if sha == "" {
			sha = side.DistSHA256
		}
		p50 := side.P50Ms
		if p50 == 0 {
			p50 = side.SourceP50
		}
		if sha == "" || side.Samples <= 0 || p50 <= 0 {
			continue
		}
		dist, err := os.ReadFile(filepath.Join(dir, name+".dist"))
		if err != nil {
			continue
		}
		sum := sha256.Sum256(dist)
		if hex.EncodeToString(sum[:]) != sha {
			continue
		}
		out = append(out, Calibration{
			Name: name, SHA256: sha, Samples: side.Samples, P50Ms: p50, Verified: true,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
