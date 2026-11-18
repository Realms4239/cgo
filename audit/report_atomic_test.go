package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteFileLeavesNoTempOnFailure(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "report.json")
	if err := atomicWriteFile(target, []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(target)
	if err != nil || string(body) != `{"ok":true}` {
		t.Fatalf("atomic write result = %q, %v", body, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("leftover temp files: %v, %v", entries, err)
	}
}
