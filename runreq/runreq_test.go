package runreq

import (
	"strings"
	"testing"
)

// La règle partagée — ici, dans le module qui en est le propriétaire.

func TestCampaignRequestValidateAllListsEveryRefusal(t *testing.T) {
	r := CampaignRequest{
		Profiles:     []string{"PX"},
		Qdiscs:       []string{"nope"},
		CCs:          nil,
		Reps:         6,
		Flows:        9,
		DeadlineMs:   5000 + 1,
		Bench:        "bga",
		Direction:    "nope",
		ChannelModel: "nope",
		Target:       strings.Repeat("x", 65),
	}
	got := r.ValidateAll()
	want := map[string]bool{
		"profiles": true, "reps": true, "flows": true, "deadline_ms": true,
		"target": true, "bench": true, "channel_model": true, "direction": true,
		"qdiscs": true, "axes": true,
	}
	for _, err := range got {
		r, ok := err.(*Refusal)
		if !ok {
			t.Fatalf("err = %v (%T), veut *Refusal", err, err)
		}
		delete(want, r.Rule)
	}
	if len(want) != 0 {
		t.Fatalf("refus manquantes : %v", want)
	}
}

func TestBurstyOnP2OrP3Allowed(t *testing.T) {
	for _, id := range []string{"P2", "P3"} {
		for _, err := range ValidateAllForTest(CampaignRequest{Profiles: []string{id}, ChannelModel: Bursty}) {
			t.Fatalf("profil %q refusé en bursty : %v", id, err)
		}
	}
	for _, id := range []string{"P1", "P4"} {
		r := CampaignRequest{Profiles: []string{id}, ChannelModel: Bursty}
		if err := r.Validate(); err == nil {
			t.Fatalf("profil %q accepté en bursty", id)
		} else if rf, ok := err.(*Refusal); !ok || !strings.Contains(rf.Msg, id) || !strings.Contains(rf.Msg, "bursty") {
			t.Fatalf("refus pour %q incomplet : %v", id, err)
		}
	}
}

// ValidateAllForTest — alias local pour éviter de coupler les tests à l'ordre des refus.
func ValidateAllForTest(r CampaignRequest) []error {
	if out := r.ValidateAll(); len(out) > 0 {
		return out
	}
	return nil
}

func TestBurstyMessageIsShared(t *testing.T) {
	m := BurstyViolationMsg("P1")
	for _, want := range []string{"P1", "bursty", "P2/P3", "clean"} {
		if !strings.Contains(m, want) {
			t.Errorf("BurstyViolationMsg(%q) = %q, manque %q", "P1", m, want)
		}
	}
}

func TestUnknownQdiscsAndCCs(t *testing.T) {
	if got := UnknownQdiscs([]string{"cake", "nope"}); len(got) != 1 || got[0] != "nope" {
		t.Fatalf("UnknownQdiscs = %v", got)
	}
	if got := UnknownCCs([]string{"cubic", "nope"}); len(got) != 1 || got[0] != "nope" {
		t.Fatalf("UnknownCCs = %v", got)
	}
	if got := UnknownQdiscs([]string{"cake", "fq_codel", "pfifo_fast"}); len(got) != 0 {
		t.Fatalf("UnknownQdiscs(%v) = %v, veut vide", []string{"cake", "fq_codel", "pfifo_fast"}, got)
	}
	if got := UnknownCCs([]string{"cubic", "bbr"}); len(got) != 0 {
		t.Fatalf("UnknownCCs(cubic,bbr) = %v, veut vide", got)
	}
}

func TestValidateOK(t *testing.T) {
	good := CampaignRequest{Profiles: []string{"P2"}, Reps: 3, Flows: 4, DeadlineMs: 1000, Bench: BenchDualNS, Direction: DirectionUp, ChannelModel: Clean}
	if err := good.Validate(); err != nil {
		t.Fatalf("request refusée : %v", err)
	}
}

func TestAxesCoFilter(t *testing.T) {
	// l'un sans l'autre = refus hors bornes — même règle que l'API
	r := CampaignRequest{Profiles: []string{"P2"}, Qdiscs: []string{"cake"}}
	if err := r.Validate(); err == nil {
		t.Fatal("qdisc seul accepté")
	} else if rf, ok := err.(*Refusal); !ok || rf.Rule != "axes" {
		t.Fatalf("refus = %v, veut rule=axes", err)
	}
	r = CampaignRequest{Profiles: []string{"P2"}, CCs: []string{"cubic"}}
	if err := r.Validate(); err == nil {
		t.Fatal("cc seul accepté")
	}
}
