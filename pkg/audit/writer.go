package audit

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Realms4239/cgo/pkg/model"
)

func AppendLinkAudit(dataDir string, r *Result) error {
	path := filepath.Join(dataDir, "link_audit.csv")
	exists := false
	if st, err := os.Stat(path); err == nil && st.Size() > 0 {
		exists = true
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if !exists {
		if err := w.Write(model.LinkAuditHeader); err != nil {
			return err
		}
	}
	row := []string{
		r.AuditID, r.Timestamp, r.Site, r.LinkType, r.Provider,
		fmt.Sprintf("%.1f", r.RTTIdleP50), fmt.Sprintf("%.1f", r.RTTIdleP95),
		fmt.Sprintf("%.1f", r.RTTLoadedP50), fmt.Sprintf("%.1f", r.RTTLoadedP95),
		fmt.Sprintf("%.1f", r.ThroughputMbps), fmt.Sprintf("%.1f", r.LossPct),
		fmt.Sprintf("%.1f", r.HTTPSmallP95), fmt.Sprintf("%.1f", r.DataUsedMB), r.Notes,
	}
	if err := w.Write(row); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}
