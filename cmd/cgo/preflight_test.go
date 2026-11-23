package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/probe"
)

// Oracle — la preuve preflight porte la source résolue et ne fabrique aucun
// hash d'objet ni compteur quand la source n'est pas ebpf.
func TestPreflightHasSourceAndNoPlaceholderHash(t *testing.T) {
	var buf bytes.Buffer
	info := probe.KernelSourceInfo{Source: "tcpinfo", Reason: "hôte non-Linux"}
	if err := writePreflight(&buf, info); err != nil {
		t.Fatalf("writePreflight: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"source"`) {
		t.Fatalf("champ source absent du preflight: %s", out)
	}

	var doc map[string]any
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("preflight JSON illisible: %v", err)
	}
	if doc["source"] != "tcpinfo" {
		t.Fatalf("source=%#v, attendu tcpinfo", doc["source"])
	}
	if v, ok := doc["object_sha256"]; ok && v != "" {
		t.Fatalf("hash d'objet fabriqué sur une source tcpinfo: %#v", v)
	}
	if v, ok := doc["reserve_failures"]; ok && v != float64(0) {
		t.Fatalf("compteur fabriqué sur une source tcpinfo: %#v", v)
	}
}

// Oracle positif — un hash réel fourni pour une source ebpf est bien rendu.
func TestPreflightCarriesRealHashForEBPF(t *testing.T) {
	var buf bytes.Buffer
	info := probe.KernelSourceInfo{Source: "ebpf", ObjectSHA256: "abc123def456"}
	if err := writePreflight(&buf, info); err != nil {
		t.Fatalf("writePreflight: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("preflight JSON illisible: %v", err)
	}
	if doc["object_sha256"] != "abc123def456" {
		t.Fatalf("hash d'objet non recopié: %#v", doc["object_sha256"])
	}
}
