package results

import (
	"path/filepath"
	"testing"
)

func TestDataRootDefaults(t *testing.T) {
	t.Setenv("CGO_DATA_DIR", "")
	if got := DataRoot(); got != "data" {
		t.Errorf("DataRoot() = %q, want data", got)
	}
	if got := RunsRoot(); got != filepath.Join("data", "runs") {
		t.Errorf("RunsRoot() = %q", got)
	}
	if got := FiguresRoot(); got != filepath.Join("data", "figures") {
		t.Errorf("FiguresRoot() = %q", got)
	}
	if got := LinkAuditCSV(); got != filepath.Join("data", "link_audit.csv") {
		t.Errorf("LinkAuditCSV() = %q", got)
	}
}

func TestDataRootPinned(t *testing.T) {
	t.Setenv("CGO_DATA_DIR", "/var/lib/cgo")
	if got := DataRoot(); got != "/var/lib/cgo" {
		t.Errorf("DataRoot() = %q", got)
	}
	if got := RunsRoot(); got != filepath.Join("/var/lib/cgo", "runs") {
		t.Errorf("RunsRoot() = %q", got)
	}
	if got := LinkAuditCSV(); got != filepath.Join("/var/lib/cgo", "link_audit.csv") {
		t.Errorf("LinkAuditCSV() = %q", got)
	}
}

func TestDataRootWhitespaceFallsBack(t *testing.T) {
	t.Setenv("CGO_DATA_DIR", "   ")
	if got := DataRoot(); got != "data" {
		t.Errorf("DataRoot() = %q, veut le défaut (espaces = absent)", got)
	}
}
