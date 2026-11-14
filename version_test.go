package cgo

import (
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// gitRevShow — blob HEAD pour un chemin. Le test verrouille le contrat
// commit/release : ce que `go test` certifie est l'arbre COMMITTÉ, pas les
// hunks humains non engagés du worktree (que cet arbre accepte en design).
func gitRevShow(t *testing.T, path string) string {
	t.Helper()
	cmd := exec.Command("git", "show", "HEAD:"+path)
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("version SSoT: HEAD indisponible (%v) — parité non prouvée sur cet hôte", err)
	}
	return string(out)
}

func TestVersionSingleSource(t *testing.T) {
	// VERSION embarqué == VERSION commité.
	version := strings.TrimSpace(gitRevShow(t, "VERSION"))
	if version == "" {
		t.Fatal("HEAD:VERSION est vide")
	}
	if Version != version {
		t.Fatalf("embed != VERSION commité: %q vs %q", Version, version)
	}
	// package.json commité.
	b := gitRevShow(t, "web/frontend/package.json")
	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal([]byte(b), &pkg); err != nil {
		t.Fatal(err)
	}
	// docs/api.md commité : version de la ligne /api/health.
	api := gitRevShow(t, "docs/api.md")
	health := ""
	for _, line := range strings.Split(api, "\n") {
		if strings.Contains(line, "/api/health") {
			health = line
			break
		}
	}
	m := regexp.MustCompile(`"version":\s*"([^"]+)"`).FindStringSubmatch(health)
	if m == nil {
		t.Fatal("pas de version dans la ligne /api/health de docs/api.md (commité)")
	}
	if pkg.Version != version || m[1] != version {
		t.Fatalf("version drift (commité): VERSION=%s package.json=%s api.md=%s", version, pkg.Version, m[1])
	}
	// Sonde octet du fichier worktree : informative uniquement, jamais
	// bloquante — le worktree peut porter du WIP humain non engagé.
	if raw, err := os.ReadFile("VERSION"); err == nil {
		if wt := strings.TrimSpace(string(raw)); wt != version {
			t.Logf("note: worktree VERSION=%s != HEAD VERSION=%s (WIP toléré, non bloquant)", wt, version)
		}
	}
}
