package results

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestSeedRunsExtracts(t *testing.T) {
	dir := t.TempDir()
	seeded, n, err := SeedRuns(dir)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if !seeded || n == 0 {
		t.Fatalf("seeded=%v n=%d, want seeded=true n>0", seeded, n)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "data", "runs", "*", "aqm_eval.csv"))
	if len(files) == 0 {
		t.Fatal("aucun aqm_eval.csv après seed")
	}
}

func TestSeedRunsNoOpWhenPresent(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, "data", "runs", "run-x")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "aqm_eval.csv"), []byte("a,b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	seeded, n, err := SeedRuns(dir)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if seeded || n != 0 {
		t.Fatalf("seeded=%v n=%d, want no-op", seeded, n)
	}
	if b, _ := os.ReadFile(filepath.Join(runDir, "aqm_eval.csv")); string(b) != "a,b\n" {
		t.Fatal("le seed a écrasé un run existant")
	}
}

func evilTar(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(zw)
	for _, name := range []string{
		"data/runs/../../evil.txt",
		"/abs/evil.txt",
		"data/runs/ok/aqm_eval.csv",
	} {
		body := []byte("x")
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(body))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.WriteHeader(&tar.Header{Name: "data/runs/link", Typeflag: tar.TypeSymlink, Linkname: "/etc/passwd"}); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestSeedRunsRejectsEvil(t *testing.T) {
	saved := runsFrozen
	runsFrozen = evilTar(t)
	defer func() { runsFrozen = saved }()
	dir := t.TempDir()
	seeded, n, err := SeedRuns(dir)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if !seeded || n != 1 {
		t.Fatalf("seeded=%v n=%d, want seeded=true n=1 (seul ok accepté)", seeded, n)
	}
	if _, err := os.Stat(filepath.Join(dir, "evil.txt")); !os.IsNotExist(err) {
		t.Fatal("traversal ../ accepté")
	}
	if _, err := os.Lstat(filepath.Join(dir, "data", "runs", "link")); !os.IsNotExist(err) {
		t.Fatal("symlink accepté")
	}
}
