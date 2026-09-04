package campagne

import (
	"context"
	"sync"
	"testing"

	"github.com/Realms4239/cgo/pkg/model"
	"github.com/Realms4239/cgo/pkg/qdisc"
)

// TestRunEventDownloadShaper — sens download : le shaper se pose sur
// l'ÉMISSION SERVEUR (veth-s via NsRunner), pas veth-c. Un runner qui
// enregistre ses appels doit montrer replace sur l'iface serveur pour la
// cellule down, et sur l'iface cliente pour la cellule up (défaut).
func TestRunEventDownloadShaper(t *testing.T) {
	var mu sync.Mutex
	var seen = map[string][]string{} // iface -> args joints

	recTC := &recordingRunner{mu: &mu, seen: seen, tag: "cli"}
	recSrv := &recordingRunner{mu: &mu, seen: seen, tag: "srv"}

	deps := fastDeps()
	deps.TC = recTC
	deps.TCShaper = recSrv
	deps.ShaperIf = "veth-s"
	deps.Small = func(context.Context) (float64, error) { return 25, nil }

	// cellule download : Deps.Direction = "down"
	d := deps
	d.Direction = "down"
	d.Bulk = func(ctx context.Context, addr string) (uint64, error) { return 2_500_000, nil }
	ev := model.Event{RunID: "rd", EventID: 1, Profile: "P2", Qdisc: model.Cake, CC: model.BBR, Repetition: 1}
	if _, err := RunEvent(context.Background(), ev, model.Profiles["P2"], d); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	if len(seen["srv:veth-s"]) == 0 {
		mu.Unlock()
		t.Fatalf("download : aucun appel shaper sur veth-s (seen=%v)", seen)
	}
	mu.Unlock()
}

// recordingRunner — TCRunner qui journalise iface/args sans exécuter tc.
type recordingRunner struct {
	mu   *sync.Mutex
	seen map[string][]string
	tag  string
}

func (r *recordingRunner) Run(args ...string) ([]byte, error) {
	if len(args) >= 4 && args[0] == "qdisc" {
		r.mu.Lock()
		r.seen[r.tag+":"+args[3]] = append(r.seen[r.tag+":"+args[3]], joinArgs(args))
		r.mu.Unlock()
	}
	return nil, nil
}

func joinArgs(args []string) string {
	out := ""
	for _, a := range args {
		out += a + " "
	}
	return out
}

var _ qdisc.TCRunner = (*recordingRunner)(nil)
