package figures

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Realms4239/cgo/pkg/results"
)

// Generate crée des figures SVG à partir des CSV gelés sous dataDir, en écrivant dans outDir.
// Il parcourt tous les runs par results.Scan, puis émet les SVG barres et nuage de points.
func Generate(dataDir, outDir string) error {
	groups, err := results.Scan(dataDir, "")
	if err != nil {
		return err
	}
	if len(groups) == 0 {
		return fmt.Errorf("no data")
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	meta := provenanceMeta(dataDir)
	// barres : small_p95 par groupe
	if err := writeBar(filepath.Join(outDir, "small_p95.svg"), groups, meta); err != nil {
		return err
	}
	if err := writeScatter(filepath.Join(outDir, "scatter.svg"), groups, meta); err != nil {
		return err
	}
	return nil
}

// ProvenanceSHA — sha256 complet du dernier aqm_eval.csv ("" sans données).
// Une seule dérivation pour la chaîne de triple provenance (Wall/Archives/Report).
func ProvenanceSHA(dataDir string) string {
	runs, _ := filepath.Glob(filepath.Join(dataDir, "*", "aqm_eval.csv"))
	if len(runs) == 0 {
		return ""
	}
	sort.Strings(runs)
	b, err := os.ReadFile(runs[len(runs)-1])
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ProvenanceHash8 — les 8 premiers caractères hex de ProvenanceSHA.
func ProvenanceHash8(dataDir string) string {
	if full := ProvenanceSHA(dataDir); full != "" {
		return full[:8]
	}
	return ""
}

func provenanceMeta(dataDir string) string {
	runs, _ := filepath.Glob(filepath.Join(dataDir, "*", "aqm_eval.csv"))
	if len(runs) == 0 {
		return ""
	}
	sort.Strings(runs)
	latest := runs[len(runs)-1]
	b, err := os.ReadFile(latest)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	sha := hex.EncodeToString(sum[:])
	src := filepath.ToSlash(latest)
	date := time.Now().UTC().Format(time.RFC3339)
	return fmt.Sprintf(`<metadata><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns:dc="http://purl.org/dc/elements/1.1/"><rdf:Description><dc:source>%s</dc:source><dc:identifier>sha256:%s</dc:identifier><dc:date>%s</dc:date><dc:creator>Meteolink</dc:creator></rdf:Description></rdf:RDF></metadata>`, src, sha, date)
}

func writeBar(path string, groups []results.Group, meta string) error {
	maxVal := 0.0
	for _, g := range groups {
		if g.Smallp95Median > maxVal {
			maxVal = g.Smallp95Median
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}
	w, h := 800, 400
	ml, mr, mt, mb := 60, 20, 40, 60
	usableW := w - ml - mr
	barW := usableW / (len(groups)*2 + 1)
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, w, h, w, h)
	svg += meta
	svg += fmt.Sprintf(`<rect width="100%%" height="100%%" fill="#070707"/><text x="%d" y="24" fill="#f2f2f4" font-family="Cormorant Garamond" font-size="16" font-weight="600">small p95 par configuration</text>`, w/2-100)
	for i, g := range groups {
		x := ml + barW + i*2*barW
		barH := int(float64(h-mt-mb) * g.Smallp95Median / maxVal)
		y := h - mb - barH
		color := "#5ad3e3"
		if g.Best {
			color = "#1fa348"
		}
		svg += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`, x, y, barW, barH, color)
		label := fmt.Sprintf("%s/%s", g.Qdisc, g.CC)
		svg += fmt.Sprintf(`<text x="%d" y="%d" fill="#8b9099" font-family="JetBrains Mono" font-size="9" text-anchor="middle">%s</text>`, x+barW/2, h-mb+14, label)
		svg += fmt.Sprintf(`<text x="%d" y="%d" fill="#f2f2f4" font-family="JetBrains Mono" font-size="10" text-anchor="middle">%.1f</text>`, x+barW/2, y-6, g.Smallp95Median)
	}
	// axes
	svg += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#26262a"/>`, ml, h-mb, w-mr, h-mb)
	svg += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#26262a"/>`, ml, mt, ml, h-mb)
	svg += `</svg>`
	return os.WriteFile(path, []byte(svg), 0644)
}

func writeScatter(path string, groups []results.Group, meta string) error {
	maxX, maxY := 0.0, 0.0
	for _, g := range groups {
		if g.GoodputMedian > maxX {
			maxX = g.GoodputMedian
		}
		if g.Smallp95Median > maxY {
			maxY = g.Smallp95Median
		}
	}
	if maxX == 0 {
		maxX = 1
	}
	if maxY == 0 {
		maxY = 1
	}
	w, h := 800, 400
	ml, mr, mt, mb := 60, 20, 40, 60
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, w, h, w, h)
	svg += meta
	svg += `<rect width="100%" height="100%" fill="#070707"/>`
	svg += fmt.Sprintf(`<text x="%d" y="24" fill="#f2f2f4" font-family="Cormorant Garamond" font-size="16" font-weight="600">compromis latence / débit</text>`, w/2-90)
	for _, g := range groups {
		x := ml + int(float64(w-ml-mr)*g.GoodputMedian/maxX)
		y := h - mb - int(float64(h-mt-mb)*g.Smallp95Median/maxY)
		color := "#b48ae0"
		if g.Best {
			color = "#1fa348"
		}
		svg += fmt.Sprintf(`<circle cx="%d" cy="%d" r="6" fill="%s" stroke="#f2f2f4" stroke-width="1"/>`, x, y, color)
		svg += fmt.Sprintf(`<text x="%d" y="%d" fill="#8b9099" font-family="JetBrains Mono" font-size="9">%s/%s</text>`, x+8, y-8, g.Qdisc, g.CC)
	}
	svg += fmt.Sprintf(`<text x="%d" y="%d" fill="#8b9099" font-family="JetBrains Mono" font-size="10" text-anchor="middle">goodput (Mbit/s)</text>`, w/2, h-8)
	svg += fmt.Sprintf(`<text x="14" y="%d" fill="#8b9099" font-family="JetBrains Mono" font-size="10" transform="rotate(-90 14 %d)">small p95 (ms)</text>`, h/2, h/2)
	svg += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#26262a"/>`, ml, h-mb, w-mr, h-mb)
	svg += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#26262a"/>`, ml, mt, ml, h-mb)
	svg += `</svg>`
	return os.WriteFile(path, []byte(svg), 0644)
}
