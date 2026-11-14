//go:build real

package qdisc

import "testing"

// TestExecRunnerRealTcReadsJSON — à exécuter dans le conteneur du banc où le
// tc réel vit (go test -tags=real). Exclu des runs par défaut (CI sans tag).
// Contract FAST TOOLS : le tag porte un test vivant, jamais un tag vide.
func TestExecRunnerRealTcReadsJSON(t *testing.T) {
	out, err := ExecRunner{}.Run("-j", "-s", "qdisc", "show", "dev", "lo")
	if err != nil {
		t.Fatalf("tc réel indisponible: %v (%s)", err, out)
	}
	stats, err := decodeStats(out)
	if err != nil {
		t.Fatalf("sortie tc réelle non décodable: %v (%s)", err, out)
	}
	if stats == nil {
		t.Fatal("decode a rendu nil sans erreur")
	}
}
