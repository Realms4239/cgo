package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func testProof() CHRProof {
	return CHRProof{
		AuditID:      "audit-1792926281",
		ReportSHA256: "aa",
		CHR:          CHRInfo{Version: "7.15.3", Arch: "x86_64", ImageSHA256: "bb"},
		Level:        "windows-tcg",
		Accelerator:  "tcg",
		Import:       ImportResult{Accepted: true},
	}
}

func TestWriteCHRProofSealsHashes(t *testing.T) {
	dir := t.TempDir()
	export := []byte("/queue type\n")
	transcript := []byte("login: admin\n")
	p := testProof()
	path, err := WriteCHRProof(dir, p, export, transcript)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got CHRProof
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Schema != "cgo.chr-proof/1" {
		t.Fatalf("schema = %q", got.Schema)
	}
	if got.ExportSHA256 != FileSHA256(export) {
		t.Fatal("export hash not sealed")
	}
	if got.TranscriptSHA256 != FileSHA256(transcript) {
		t.Fatal("transcript hash not sealed")
	}
	for _, name := range []string{"export.txt", "transcript.txt"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("%s not written: %v", name, err)
		}
	}
}

func TestWriteCHRProofRejectsUnknownLevel(t *testing.T) {
	if _, err := WriteCHRProof(t.TempDir(), CHRProof{AuditID: "a", Level: "kvm"}, nil, nil); err == nil {
		t.Fatal("unknown level accepted")
	}
}

func TestWriteCHRProofRejectsEmptyAuditID(t *testing.T) {
	if _, err := WriteCHRProof(t.TempDir(), CHRProof{Level: "windows-tcg"}, nil, nil); err == nil {
		t.Fatal("empty audit_id accepted")
	}
}
