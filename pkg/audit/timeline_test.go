package audit

import "testing"

func TestTimelineBucketsBySecondAndKeepsGaps(t *testing.T) {
	tl := newTimeline()
	// 3 s de repos, 2 s de charge sans réponse HTTP, puis une réponse.
	for i := 0; i < 15; i++ {
		tl.add(i/5, "baseline", 24+float64(i), 0, false)
	}
	tl.add(3, "charge", 300, 0, false)
	tl.add(4, "charge", 320, 900, true)

	got := tl.points()
	if len(got) != 5 {
		t.Fatalf("points = %d, voulu 5 (%+v)", len(got), got)
	}
	if got[0].N != 5 {
		t.Fatalf("n seconde 0 = %d, voulu 5", got[0].N)
	}
	if got[3].SmallP95 != nil {
		t.Fatal("une sonde muette ne doit pas devenir une mesure")
	}
	if got[4].SmallP95 == nil || *got[4].SmallP95 != 900 {
		t.Fatalf("small_p95 = %v, voulu 900", got[4].SmallP95)
	}
	if got[3].RTTP95 == nil {
		t.Fatal("le RTT de la seconde 3 a pourtant répondu")
	}
}

func TestTimelineCapIsTotalPlusFive(t *testing.T) {
	tl := newTimeline()
	for i := 0; i < 200; i++ {
		tl.add(i, "charge", 50, 0, false)
	}
	if len(tl.points()) > 65 {
		t.Fatalf("pas de garde de longueur : %d points", len(tl.points()))
	}
}
