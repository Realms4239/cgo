package audit

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Realms4239/cgo/pkg/model"
)

func AppendLinkAudit(dataDir string, r *Result) error {
	path := filepath.Join(dataDir, "link_audit.csv")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	newCols := []string{"bloat_delta_ms", "bloat_grade", "bloat_verdict", "profile_match", "match_delta"}
	fullHdr := append(append([]string{}, model.LinkAuditHeader...), newCols...)
	// migration de schéma : un historique aux colonnes courtes (14, puis 17)
	// reçoit les colonnes manquantes rembourrées — pas de CSV bancal, les
	// lecteurs par position (ReadAll verrouille au premier enregistrement)
	// ne cassent pas au milieu du fichier.
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		rd := csv.NewReader(bytes.NewReader(data))
		rd.FieldsPerRecord = -1
		rows, err := rd.ReadAll()
		if err == nil && len(rows) > 0 && len(rows[0]) < len(fullHdr) {
			var buf bytes.Buffer
			wr := csv.NewWriter(&buf)
			_ = wr.Write(fullHdr)
			for _, row := range rows[1:] {
				for len(row) < len(fullHdr) {
					row = append(row, "")
				}
				_ = wr.Write(row)
			}
			wr.Flush()
			if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
				return err
			}
		}
	}
	exists := false
	if st, err := os.Stat(path); err == nil && st.Size() > 0 {
		exists = true
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if !exists {
		// en-tête complet : bloat + rapprochement en fin de ligne — les
		// lecteurs par nom tolèrent l'absence sur l'ancien historique, et
		// la migration ci-dessus a déjà rembourré les vieux fichiers
		if err := w.Write(fullHdr); err != nil {
			return err
		}
	}
	row := []string{
		r.AuditID, r.Timestamp, r.Site, r.LinkType, r.Provider,
		fmt.Sprintf("%.1f", r.RTTIdleP50), fmt.Sprintf("%.1f", r.RTTIdleP95),
		fmt.Sprintf("%.1f", r.RTTLoadedP50), fmt.Sprintf("%.1f", r.RTTLoadedP95),
		fmt.Sprintf("%.1f", r.ThroughputMbps), fmt.Sprintf("%.1f", r.LossPct),
		fmt.Sprintf("%.1f", r.HTTPSmallP95), fmt.Sprintf("%.1f", r.DataUsedMB), r.Notes,
		fmt.Sprintf("%.1f", r.BloatDeltaMs), r.BloatGrade, r.BloatVerdict,
		r.ProfileMatch, r.MatchDelta,
	}
	if err := w.Write(row); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}
