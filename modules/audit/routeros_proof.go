package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type CHRInfo struct {
	Version     string `json:"version"`
	Arch        string `json:"arch"`
	ImageSHA256 string `json:"image_sha256"`
}

type ImportResult struct {
	Accepted  bool   `json:"accepted"`
	ErrorLine string `json:"error_line,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// CHRProof is the sealed evidence that a RouterOS instance accepted the
// generated configuration. It is written under data/ (never committed) and
// cited from docs/VERIFY.md by path and hash.
type CHRProof struct {
	Schema           string          `json:"schema"`
	AuditID          string          `json:"audit_id"`
	ReportSHA256     string          `json:"report_sha256"`
	CHR              CHRInfo         `json:"chr"`
	Level            string          `json:"level"`
	Accelerator      string          `json:"accelerator"`
	Import           ImportResult    `json:"import"`
	Checks           []RouterOSCheck `json:"checks"`
	ExportSHA256     string          `json:"export_sha256"`
	TranscriptSHA256 string          `json:"transcript_sha256"`
}

func FileSHA256(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// WriteCHRProof writes proof.json, export.txt and transcript.txt atomically
// (temp file + rename) so a killed run never leaves a half proof that reads
// as complete.
func WriteCHRProof(dir string, p CHRProof, export, transcript []byte) (string, error) {
	if p.AuditID == "" {
		return "", fmt.Errorf("chr: audit_id absent")
	}
	if p.Level != "windows-tcg" && p.Level != "linux-namespace" {
		return "", fmt.Errorf("chr: niveau %q inconnu (windows-tcg ou linux-namespace)", p.Level)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	p.Schema = "cgo.chr-proof/1"
	p.ExportSHA256 = FileSHA256(export)
	p.TranscriptSHA256 = FileSHA256(transcript)
	if err := writeAtomic(filepath.Join(dir, "export.txt"), export); err != nil {
		return "", err
	}
	if err := writeAtomic(filepath.Join(dir, "transcript.txt"), transcript); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "proof.json")
	if err := writeAtomic(path, b); err != nil {
		return "", err
	}
	return path, nil
}

func writeAtomic(path string, b []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
