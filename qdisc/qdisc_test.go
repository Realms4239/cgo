package qdisc

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/Realms4239/cgo/model"
)

func TestApplyShaperArgVectors(t *testing.T) {
	r := &FakeRunner{}
	if err := ApplyShaper(r, "veth-s", model.Cake, 20, 100); err != nil {
		t.Fatal(err)
	}
	if err := ApplyShaper(r, "veth-s", model.FqCodel, 20, 0); err != nil {
		t.Fatal(err)
	}
	if err := ApplyShaper(r, "veth-s", model.PfifoFast, 80, 0); err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"qdisc", "replace", "dev", "veth-s", "parent", "1:", "handle", "10:", "cake", "bandwidth", "20mbit", "rtt", "100ms",
			"ack-filter", "nat", "diffserv4", "triple-isolate"},
		{"qdisc", "replace", "dev", "veth-s", "parent", "1:", "handle", "10:", "tbf", "rate", "20mbit", "burst", "25000", "latency", "400ms"},
		{"qdisc", "replace", "dev", "veth-s", "parent", "10:1", "handle", "20:", "fq_codel", "target", "5ms", "interval", "100ms", "limit", "10240", "quantum", "1514", "flows", "1024"},
		{"qdisc", "replace", "dev", "veth-s", "parent", "1:", "handle", "10:", "tbf", "rate", "80mbit", "burst", "100000", "latency", "400ms"},
	}
	if !reflect.DeepEqual(r.Calls, want) {
		t.Fatalf("got %v\nwant %v", r.Calls, want)
	}
}

// scriptRunner rejoue une séquence fixe d'issues (succès ou erreur) et
// capture les appels — pour les branches de repli racine d'ApplyShaper.
type scriptRunner struct {
	script []error // un élément par appel ; nil = succès
	calls  [][]string
}

func (s *scriptRunner) Run(args ...string) ([]byte, error) {
	i := len(s.calls)
	s.calls = append(s.calls, args)
	if i < len(s.script) && s.script[i] != nil {
		return []byte("simulated tc failure"), s.script[i]
	}
	return nil, nil
}

// T1 : cake parent 1: échoue → cake racine réussit → nil ; 2 appels.
func TestShaperCakeRootFallbackSuccess(t *testing.T) {
	r := &scriptRunner{script: []error{errors.New("Cannot find specified qdisc")}}
	if err := ApplyShaper(r, "e-srv", model.Cake, 20, 100); err != nil {
		t.Fatalf("root fallback should succeed: %v", err)
	}
	if len(r.calls) != 2 {
		t.Fatalf("want 2 calls, got %d: %v", len(r.calls), r.calls)
	}
	if !strings.Contains(strings.Join(r.calls[1], " "), "root handle 1: cake") {
		t.Fatalf("second call is not the root cake: %v", r.calls[1])
	}
}

// T2 : cake parent 1: échoue → cake racine échoue → erreur nommant les deux.
func TestShaperCakeRootFallbackError(t *testing.T) {
	r := &scriptRunner{script: []error{errors.New("parent fail"), errors.New("root fail")}}
	err := ApplyShaper(r, "e-srv", model.Cake, 20, 100)
	if err == nil {
		t.Fatal("double failure must return an error")
	}
	if len(r.calls) != 2 {
		t.Fatalf("want 2 calls, got %d: %v", len(r.calls), r.calls)
	}
	msg := err.Error()
	if !strings.Contains(msg, "parent") || !strings.Contains(msg, "root") {
		t.Fatalf("error must name both attempts: %q", msg)
	}
}

// T3 : tbf parent échoue → tbf racine réussit → fq_codel enfant 1:1 réussit
// → nil ; 3 appels, le 3e porte parent 1:1 et les params RFC 8290 épinglés.
func TestShaperFqCodelRootFallbackAttachesChild(t *testing.T) {
	r := &scriptRunner{script: []error{errors.New("Cannot find specified qdisc")}}
	if err := ApplyShaper(r, "e-srv", model.FqCodel, 20, 0); err != nil {
		t.Fatalf("root fallback should succeed: %v", err)
	}
	if len(r.calls) != 3 {
		t.Fatalf("want 3 calls (tbf parent, tbf root, fq_codel child), got %d: %v", len(r.calls), r.calls)
	}
	third := strings.Join(r.calls[2], " ")
	want := "qdisc replace dev e-srv parent 1:1 handle 20: fq_codel target 5ms interval 100ms limit 10240 quantum 1514 flows 1024"
	if third != want {
		t.Fatalf("third call:\n got %q\nwant %q", third, want)
	}
}

// T4 : tbf parent échoue → tbf racine échoue → erreur ; 2 appels seulement
// (l'enfant n'est jamais attaché sans shaper).
func TestShaperFqCodelRootFallbackError(t *testing.T) {
	r := &scriptRunner{script: []error{errors.New("parent fail"), errors.New("root fail")}}
	if err := ApplyShaper(r, "e-srv", model.FqCodel, 20, 0); err == nil {
		t.Fatal("double tbf failure must return an error")
	}
	if len(r.calls) != 2 {
		t.Fatalf("want 2 calls, got %d: %v", len(r.calls), r.calls)
	}
}

func TestApplyNetemLossOmittedWhenZero(t *testing.T) {
	r := &FakeRunner{}
	if err := ApplyNetem(r, "veth-c", 100, 15, 0.5); err != nil {
		t.Fatal(err)
	}
	if err := ApplyNetem(r, "veth-c", 20, 2, 0); err != nil {
		t.Fatal(err)
	}
	got := r.Calls
	if len(got) != 2 || got[0][11] != "loss" || len(got[1]) != 11 {
		t.Fatalf("unexpected vectors: %v", got)
	}
}

func TestFQCoDel_PinsRFC8290Params(t *testing.T) {
	s := fqCoDelString("eth0")
	for _, want := range []string{"target 5ms", "interval 100ms", "limit 10240", "quantum 1514", "flows 1024"} {
		if !contains(s, want) {
			t.Fatalf("missing %q in %q", want, s)
		}
	}
}
func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
