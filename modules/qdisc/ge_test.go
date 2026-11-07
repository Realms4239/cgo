package qdisc

import (
	"strings"
	"testing"
)

func TestGEMeanLossMatchesPaperModel(t *testing.T) {
	// Exemple travaillé de l'article des implémenteurs (Carry/Briscoe,
	// netem) : p=0.01, r=0.19, 1-h=0.81, 1-k=0.01 → 5 % de perte moyenne
	// en salves. La formule : π_mauvais = p/(p+r) ; perte = π·(1-h) + (1-π)·(1-k).
	got := GEMeanLoss(GEModel{GoodToBad: 0.01, BadToGood: 0.19, LossBad: 0.81, LossGood: 0.01})
	if got < 4.9 || got > 5.1 {
		t.Fatalf("exemple article : %v %% (attendu ≈5)", got)
	}
	// Paramètres P2 du banc durci : ≈0.544 %.
	p2 := GEMeanLoss(GEModel{GoodToBad: 0.001, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.0005})
	if p2 < 0.5 || p2 > 0.6 {
		t.Fatalf("P2 : %v %% (attendu ≈0.544)", p2)
	}
	// P3 : ≈1.08 %.
	p3 := GEMeanLoss(GEModel{GoodToBad: 0.002, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.001})
	if p3 < 1.0 || p3 > 1.15 {
		t.Fatalf("P3 : %v %% (attendu ≈1.08)", p3)
	}
	if GEMeanLoss(GEModel{}) != 0 {
		t.Fatal("modèle vide : 0, pas une division par zéro")
	}
}

func TestApplyNetemGEFourArguments(t *testing.T) {
	// LA correction centrale : la forme documentée est QUATRE arguments
	// « p r 1-h 1-k » — l'ancienne forme à cinq (p r h 1-r k) est invalide
	// sur Linux et disparaît.
	f := &FakeRunner{}
	ge := GEModel{GoodToBad: 0.001, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.0005}
	if err := ApplyNetemGE(f, "w-edge", 100, 15, ge); err != nil {
		t.Fatal(err)
	}
	call := f.Calls[0]
	idx := -1
	for i, a := range call {
		if a == "gemodel" {
			idx = i
			break
		}
	}
	if idx < 0 || len(call) != idx+5 {
		t.Fatalf("gemodel doit recevoir EXACTEMENT 4 args : %v", call)
	}
	want := []string{"0.001", "0.1", "0.5", "0.0005"}
	for i, w := range want {
		if call[idx+1+i] != w {
			t.Fatalf("arg %d = %q, want %q (ordre p, r, 1-h, 1-k)", i+1, call[idx+1+i], w)
		}
	}
}

func TestApplyNetemPhysicsUniformUnchanged(t *testing.T) {
	// ge nil ⇒ la forme uniforme historique, à l'octet près — la
	// reproduction du corpus gelé en dépend.
	f := &FakeRunner{}
	if err := ApplyNetemPhysics(f, "w-edge", 100, 15, 0.5, nil); err != nil {
		t.Fatal(err)
	}
	if joined := strings.Join(f.Calls[0], " "); !strings.Contains(joined, "loss 0.5%") {
		t.Fatalf("forme uniforme attendue : %v", f.Calls[0])
	}
}

func TestApplyNetemPhysicsBursty(t *testing.T) {
	// ge fourni ⇒ gemodel 4 args, PAS de « loss X% » uniforme en plus.
	f := &FakeRunner{}
	ge := GEModel{GoodToBad: 0.002, BadToGood: 0.1, LossBad: 0.5, LossGood: 0.001}
	if err := ApplyNetemPhysics(f, "c-wan", 600, 30, 1, &ge); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(f.Calls[0], " ")
	if !strings.Contains(joined, "loss gemodel 0.002 0.1 0.5 0.001") {
		t.Fatalf("gemodel 4 args attendu : %v", f.Calls[0])
	}
	if strings.Contains(joined, "1%") {
		t.Fatalf("perte uniforme ne doit pas accompagner gemodel : %v", f.Calls[0])
	}
}
