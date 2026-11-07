package audit

import "github.com/Realms4239/cgo/metrics"

// TimelinePoint — une seconde de la fenêtre d'audit. Les valeurs sont pointeurs
// : nil veut dire « aucune sonde n'a répondu », ce qui n'est pas un 0 ms.
type TimelinePoint struct {
	T        int      `json:"t"`
	Phase    string   `json:"phase"`
	N        int      `json:"n"`
	RTTP50   *float64 `json:"rtt_p50"`
	RTTP95   *float64 `json:"rtt_p95"`
	SmallP95 *float64 `json:"small_p95"`
}

const timelineCapSec = 65

type secBucket struct {
	phase  string
	rtts   []float64
	smalls []float64
}

type timeline struct {
	buckets map[int]*secBucket
	max     int
	capSec  int
}

func newTimeline() *timeline {
	return &timeline{buckets: map[int]*secBucket{}, capSec: timelineCapSec}
}

// newTimelineFor — la fenêtre d'un run de totalSec secondes : le plafond suit
// la durée (total_sec + 5). Un run de 120 s garde sa courbe entière ; le
// plafond fixe de 65 s ne valait que pour le run par défaut de 60 s.
func newTimelineFor(totalSec int) *timeline {
	capSec := totalSec + 5
	if capSec < 1 {
		capSec = timelineCapSec
	}
	return &timeline{buckets: map[int]*secBucket{}, capSec: capSec}
}

// add enregistre une sonde. smallMs <= 0 ou smallOK faux = sonde muette.
func (t *timeline) add(sec int, phase string, rttMs, smallMs float64, smallOK bool) {
	if sec < 0 || sec >= t.capSec {
		return
	}
	b := t.buckets[sec]
	if b == nil {
		b = &secBucket{}
		t.buckets[sec] = b
	}
	if b.phase == "" {
		b.phase = phase
	}
	if rttMs > 0 {
		b.rtts = append(b.rtts, rttMs)
	}
	if smallOK && smallMs > 0 {
		b.smalls = append(b.smalls, smallMs)
	}
	if sec > t.max {
		t.max = sec
	}
}

// points rend une entrée par seconde écoulée, les secondes muettes comprises :
// un trou est une seconde où rien n'a répondu (n=0, valeurs nil), pas un index
// absent que le client devrait deviner.
func (t *timeline) points() []TimelinePoint {
	if len(t.buckets) == 0 {
		return nil
	}
	out := make([]TimelinePoint, 0, t.max+1)
	for s := 0; s <= t.max; s++ {
		p := TimelinePoint{T: s}
		if b := t.buckets[s]; b != nil {
			p.Phase = b.phase
			p.N = len(b.rtts)
			if len(b.rtts) > 0 {
				sum := summarize(b.rtts)
				p.RTTP50, p.RTTP95 = &sum.P50, &sum.P95
			}
			if len(b.smalls) > 0 {
				sum := summarize(b.smalls)
				p.SmallP95 = &sum.P95
			}
		}
		out = append(out, p)
	}
	return out
}

func summarize(xs []float64) metrics.Summary { return metrics.Summarize(xs) }
