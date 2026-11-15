package audit

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Realms4239/cgo/model"
)

// atomicWriteFile — écriture insécable (tmp + rename même FS, fsync du
// fichier avant le rename) : une migration de schéma interrompue ne laisse
// jamais un link_audit.csv tronqué derrière elle — les lecteurs par position
// verrouillent au premier enregistrement, un fichier à moitié écrit casse
// toute la reprise.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".cgo-audit-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}

func AppendLinkAudit(dataDir string, r *Result) error {
	path := filepath.Join(dataDir, "link_audit.csv")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	newCols := []string{"bloat_delta_ms", "bloat_grade", "bloat_verdict", "profile_match", "match_delta", "cap_deadline_ok_pct", "cap_sla_breach"}
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
			if err := atomicWriteFile(path, buf.Bytes(), 0644); err != nil {
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
	// CAP non mesurée : colonnes vides — une absence, pas un zéro qui se
	// lirait « 0 % sous échéance, conforme ».
	capPct, capBreach := "", ""
	if r.CAPSLAMeasured {
		capPct = fmt.Sprintf("%.1f", r.CAPDeadlineOKPct)
		capBreach = fmt.Sprintf("%t", r.CAPSLABreach)
	}
	row := []string{
		r.AuditID, r.Timestamp, r.Site, r.LinkType, r.Provider,
		fmt.Sprintf("%.1f", r.RTTIdleP50), fmt.Sprintf("%.1f", r.RTTIdleP95),
		fmt.Sprintf("%.1f", r.RTTLoadedP50), fmt.Sprintf("%.1f", r.RTTLoadedP95),
		fmt.Sprintf("%.1f", r.ThroughputMbps), fmt.Sprintf("%.1f", r.LossPct),
		fmt.Sprintf("%.1f", r.HTTPSmallP95), fmt.Sprintf("%.1f", r.DataUsedMB), r.Notes,
		fmt.Sprintf("%.1f", r.BloatDeltaMs), r.BloatGrade, r.BloatVerdict,
		r.ProfileMatch, r.MatchDelta,
		capPct, capBreach,
	}
	if err := w.Write(row); err != nil {
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	// la ligne gelée d'abord sur le disque, le seal ensuite ailleurs : le
	// sceau doit pouvoir prouver EXACTEMENT ce fichier — fsync explicite
	// avant le Close différé.
	return f.Sync()
}
