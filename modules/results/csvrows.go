package results

import (
	"encoding/csv"
	"os"
	"strings"
)

// ReadAQM lit un aqm_eval.csv gelé en tolérant les schémas historiques :
// 17 colonnes (pré-qdi_ms), 18 (qdi_ms sans voip_r), 19 (actuel, +voip_r).
// Chaque fichier porte son en-tête ; l'accès se fait par nom via ColIndex,
// jamais par position — l'insertion de qdi_ms en 8 avait décalé tous les
// lecteurs positionnels (Scan/delta/replay servaient les colonnes voisines :
// QDI étiqueté small_p95, deadline_ok_pct étiqueté goodput, cpu_pct testé
// comme gate_status ⇒ quarantaine toujours 0).
// Les lignes vides sont sautées ; les lignes plus courtes que l'en-tête
// gardent leurs champs présents (champ manquant = absent, pas d'erreur) ;
// un fichier entier n'est plus jeté pour une ligne abîmée.
func ReadAQM(path string) (header []string, rows [][]string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	all, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(all) == 0 {
		return nil, nil, nil
	}
	header = all[0]
	for _, row := range all[1:] {
		if len(row) == 0 {
			continue
		}
		if len(row) == 1 && strings.TrimSpace(row[0]) == "" {
			continue
		}
		rows = append(rows, row)
	}
	return header, rows, nil
}

// ColIndex rend l'index de la colonne nommée dans l'en-tête, -1 si absente
// (schéma historique sans cette colonne).
func ColIndex(header []string, name string) int {
	for i, h := range header {
		if strings.TrimSpace(h) == name {
			return i
		}
	}
	return -1
}

// Field lit la colonne nommée d'une ligne ("" si absente ou hors limites).
func Field(header []string, row []string, name string) string {
	i := ColIndex(header, name)
	if i < 0 || i >= len(row) {
		return ""
	}
	return row[i]
}
