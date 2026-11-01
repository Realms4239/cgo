package remediate

import (
	"math"
	"strings"
	"testing"
)

func goodLink() Link {
	return Link{
		Interface:         "wan0",
		RouterOSInterface: "ether1-wan",
		CapacityMbps:      20,
		RTTms:             100,
		Overhead:          0,
		MTU:               1500,
	}
}

// The whole point of the package: no measurement, no config. A remediation
// built from a guessed bandwidth gets pasted into a production router.
func TestRefusesUnmeasuredCapacity(t *testing.T) {
	l := goodLink()
	l.CapacityMbps = 0
	if _, err := LinuxTC(l, QdiscCake); err == nil {
		t.Fatal("emitted tc config for a zero-capacity link")
	}
	if _, err := RouterOS(l, QdiscCake); err == nil {
		t.Fatal("emitted RouterOS config for a zero-capacity link")
	}
}

func TestRefusesUnmeasuredRTT(t *testing.T) {
	l := goodLink()
	l.RTTms = 0
	if _, err := LinuxTC(l, QdiscCake); err == nil {
		t.Fatal("emitted tc config with no measured RTT")
	}
}

func TestRefusesNegativeCapacity(t *testing.T) {
	l := goodLink()
	l.CapacityMbps = -5
	if err := l.Validate(); err == nil {
		t.Fatal("accepted a negative capacity")
	}
}

func TestRefusesMissingInterface(t *testing.T) {
	l := goodLink()
	l.Interface = ""
	if _, err := LinuxTC(l, QdiscCake); err == nil {
		t.Fatal("emitted tc config with no interface")
	}
	l = goodLink()
	l.RouterOSInterface = ""
	if _, err := RouterOS(l, QdiscCake); err == nil {
		t.Fatal("emitted RouterOS config with no interface")
	}
}

func TestUnknownQdiscRejected(t *testing.T) {
	if _, err := LinuxTC(goodLink(), "hfsc"); err == nil {
		t.Fatal("accepted an unknown qdisc")
	}
}

// CAKE must be the root qdisc: stacking it under a shaper reintroduces the
// bufferbloat this project exists to remove.
func TestCakeIsRootNotChild(t *testing.T) {
	cmds, err := LinuxTC(goodLink(), QdiscCake)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(cmds, "\n")
	if !strings.Contains(joined, "root cake") {
		t.Fatalf("cake is not root:\n%s", joined)
	}
	if strings.Contains(joined, "parent 1:") {
		t.Fatalf("cake is stacked under a parent, reintroducing the queue:\n%s", joined)
	}
	// The knobs that make CAKE worth deploying on this link.
	for _, want := range []string{"ack-filter", "nat", "diffserv4", "triple-isolate"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in cake config:\n%s", want, joined)
		}
	}
}

// fq_codel has no shaper, so it must sit under a tbf whose rate is the
// measured capacity — and the tbf must not become the buffer.
func TestFqCodelUnderTbfWithMeasuredRate(t *testing.T) {
	cmds, err := LinuxTC(goodLink(), QdiscFqCodel)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(cmds, "\n")
	if !strings.Contains(joined, "tbf rate 19mbit") {
		t.Fatalf("tbf rate is not the shaped (95%%) capacity:\n%s", joined)
	}
	if !strings.Contains(joined, "latency 400ms") {
		t.Fatalf("tbf has no latency cap, so it can itself bufferbloat:\n%s", joined)
	}
	if !strings.Contains(joined, "fq_codel") {
		t.Fatalf("no fq_codel child:\n%s", joined)
	}
	// Ordering: the shaper must exist before the child attaches.
	if strings.Index(joined, "tbf") > strings.Index(joined, "fq_codel") {
		t.Fatalf("fq_codel is attached before the shaper exists:\n%s", joined)
	}
}

func TestOverheadOnlyWhenSet(t *testing.T) {
	l := goodLink()
	cmds, _ := LinuxTC(l, QdiscCake)
	if strings.Contains(strings.Join(cmds, "\n"), "overhead") {
		t.Fatal("emitted overhead when none was measured")
	}
	l.Overhead = 38
	cmds, _ = LinuxTC(l, QdiscCake)
	if !strings.Contains(strings.Join(cmds, "\n"), "overhead 38") {
		t.Fatal("dropped a measured overhead")
	}
}

func TestRouterOSCakeSizedInBits(t *testing.T) {
	script, err := RouterOS(goodLink(), QdiscCake)
	if err != nil {
		t.Fatal(err)
	}
	// 20 Mbit/s measured, shaped at 95% = 19000000 bit/s. A shaper at exactly
	// the measured rate fights the real bottleneck and under-delivers.
	if !strings.Contains(script, "max-limit=19000000") {
		t.Fatalf("RouterOS limit is not the shaped rate in bits:\n%s", script)
	}
	if !strings.Contains(script, "cake-bandwidth=19000000") {
		t.Fatalf("cake-bandwidth is not the shaped rate:\n%s", script)
	}
	// The measured rate must still be stated, so the operator can see what the
	// shaper was derived from.
	if !strings.Contains(script, "20.00 Mbit/s") {
		t.Fatalf("script does not state the measured bottleneck:\n%s", script)
	}
	for _, want := range []string{"cake-flowmode=triple-isolate", "cake-nat=yes", "cake-diffserv=diffserv4"} {
		if !strings.Contains(script, want) {
			t.Fatalf("missing %q:\n%s", want, script)
		}
	}
	if !strings.Contains(script, "parent=ether1-wan") {
		t.Fatalf("queue tree is not attached to the measured interface:\n%s", script)
	}
}

// The shaper must sit below the measured bottleneck, never at or above it.
func TestShaperBelowMeasuredRate(t *testing.T) {
	cmds, err := LinuxTC(goodLink(), QdiscCake)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(cmds, "\n")
	if strings.Contains(joined, "bandwidth 20mbit") {
		t.Fatalf("shaper set at the measured rate instead of below it:\n%s", joined)
	}
	if !strings.Contains(joined, "bandwidth 19mbit") {
		t.Fatalf("shaper is not 95%% of measured:\n%s", joined)
	}
}

// RouterOS has no fq_codel. The script must say so instead of silently
// shipping sfq as if it were equivalent.
func TestRouterOSFqCodelIsHonest(t *testing.T) {
	script, err := RouterOS(goodLink(), QdiscFqCodel)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "no fq_codel") {
		t.Fatalf("did not disclose that RouterOS lacks fq_codel:\n%s", script)
	}
	if !strings.Contains(script, "NOT fq_codel") {
		t.Fatalf("did not distinguish sfq from fq_codel:\n%s", script)
	}
}

func TestRouterOSIsIdempotent(t *testing.T) {
	script, _ := RouterOS(goodLink(), QdiscCake)
	if !strings.Contains(script, "remove [find where comment=\"cgo-managed\"]") {
		t.Fatal("script is not idempotent: a re-run would stack queues")
	}
}

func TestRecommendAsymmetricPrefersCake(t *testing.T) {
	if r := Recommend(goodLink(), true, false); r.Qdisc != QdiscCake {
		t.Fatalf("asymmetric link should get cake, got %s", r.Qdisc)
	}
	if r := Recommend(goodLink(), true, true); r.Qdisc != QdiscCake {
		t.Fatalf("asymmetric link should get cake even when CPU-bound, got %s", r.Qdisc)
	}
}

func TestRecommendSymmetricCPUConstrainedPrefersFqCodel(t *testing.T) {
	if r := Recommend(goodLink(), false, true); r.Qdisc != QdiscFqCodel {
		t.Fatalf("symmetric CPU-bound link should get fq_codel, got %s", r.Qdisc)
	}
}

func TestRecommendAlwaysJustified(t *testing.T) {
	for _, tc := range []struct{ asym, cpu bool }{{true, false}, {false, true}, {false, false}} {
		if r := Recommend(goodLink(), tc.asym, tc.cpu); r.Reason == "" {
			t.Fatalf("recommendation without a reason: %+v", r)
		}
	}
}

func TestRouterOSIngressAssumesBridgeAndSaysSo(t *testing.T) {
	l := Link{RouterOSInterface: "ether1-wan", CapacityMbps: 18.5, RTTms: 100}
	script, err := RouterOSIngress(l, QdiscCake)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "parent=bridge") {
		t.Fatalf("script sans interface par défaut :\n%s", script)
	}
	if !strings.Contains(script, "ATTENTION") {
		t.Fatalf("l'avertissement d'interface supposée est absent :\n%s", script)
	}
}

func TestRouterOSIngressUsesDeclaredLanInterface(t *testing.T) {
	l := Link{RouterOSInterface: "ether1-wan", RouterOSLanInterface: "lan-bridge", CapacityMbps: 18.5, RTTms: 100}
	script, err := RouterOSIngress(l, QdiscCake)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "parent=lan-bridge") {
		t.Fatalf("interface déclarée ignorée :\n%s", script)
	}
	if strings.Contains(script, "ATTENTION") {
		t.Fatal("aucune hypothèse à signaler quand l'interface est déclarée")
	}
}

func TestShapedRateMatchesHeadroom(t *testing.T) {
	if got := ShapedRate(18.5); math.Abs(got-17.575) > 1e-9 {
		t.Fatalf("ShapedRate(18.5) = %v, voulu 17.575", got)
	}
}
