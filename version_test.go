package cgo

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestVersionSingleSource(t *testing.T) {
	raw, err := os.ReadFile("VERSION")
	if err != nil {
		t.Fatal(err)
	}
	embedded := Version
	if raw := strings.TrimSpace(string(raw)); raw != embedded {
		t.Fatalf("embed != VERSION file: %q vs %q", embedded, raw)
	}
	b, err := os.ReadFile("web/frontend/package.json")
	if err != nil {
		t.Fatal(err)
	}
	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(b, &pkg); err != nil {
		t.Fatal(err)
	}
	api, err := os.ReadFile("docs/api.md")
	if err != nil {
		t.Fatal(err)
	}
	health := ""
	for _, line := range strings.Split(string(api), "\n") {
		if strings.Contains(line, "/api/health") {
			health = line
			break
		}
	}
	m := regexp.MustCompile(`"version":\s*"([^"]+)"`).FindStringSubmatch(health)
	if m == nil {
		t.Fatal("no version in the /api/health row of docs/api.md")
	}
	if pkg.Version != embedded || m[1] != embedded {
		t.Fatalf("version drift: VERSION=%s package.json=%s api.md=%s", embedded, pkg.Version, m[1])
	}
}
