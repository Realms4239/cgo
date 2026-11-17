package audit

import (
	"strings"
	"testing"
)

const exportCake = `/queue type
add cake-ack-filter=filter cake-bandwidth=19000000 cake-diffserv=diffserv4 \
    cake-flowmode=triple-isolate cake-nat=yes cake-rtt=25ms cake-wash=yes \
    kind=cake name=cgo-cake
/queue tree
add comment=cgo-managed max-limit=19000000 name=cgo-wan-out parent=ether1-wan \
    queue=cgo-cake
`

func TestCompareMatchPasses(t *testing.T) {
	checks := CompareRouterOS(ParseRouterOSConfig(scriptCake), ParseRouterOSConfig(exportCake))
	if len(checks) == 0 {
		t.Fatal("no checks produced")
	}
	for _, c := range checks {
		if !c.Pass {
			t.Fatalf("%s failed: expected %q found %q", c.Name, c.Expected, c.Found)
		}
	}
}

func TestCompareMutatedExportFails(t *testing.T) {
	mutated := strings.Replace(exportCake, "cake-ack-filter=filter", "cake-ack-filter=none", 1)
	checks := CompareRouterOS(ParseRouterOSConfig(scriptCake), ParseRouterOSConfig(mutated))
	var failed []RouterOSCheck
	for _, c := range checks {
		if !c.Pass {
			failed = append(failed, c)
		}
	}
	if len(failed) != 1 {
		t.Fatalf("%d failures, want 1: %+v", len(failed), failed)
	}
	if got := failed[0].Found; got != "none" {
		t.Fatalf("found = %q, want none", got)
	}
}

func TestCompareMissingObjectFails(t *testing.T) {
	checks := CompareRouterOS(ParseRouterOSConfig(scriptCake), ParseRouterOSConfig("/queue type\n"))
	var sawMissing bool
	for _, c := range checks {
		if !c.Pass && c.Found == "absent" {
			sawMissing = true
		}
	}
	if !sawMissing {
		t.Fatal("missing queue type reported as present")
	}
}

func TestCompareDuplicateObjectFails(t *testing.T) {
	dup := exportCake + "/queue type\nadd kind=cake name=cgo-cake\n"
	checks := CompareRouterOS(ParseRouterOSConfig(scriptCake), ParseRouterOSConfig(dup))
	var sawDup bool
	for _, c := range checks {
		if !c.Pass && strings.Contains(c.Name, "duplicate") {
			sawDup = true
		}
	}
	if !sawDup {
		t.Fatal("duplicate object not reported")
	}
}

func TestNormalizeROSComparesRatesSemantically(t *testing.T) {
	pairs := [][2]string{
		{"19000000", "19M"},
		{"19000000", "19.0Mbps"},
		{"19000000", "19.000000Mbps"},
		{"19000000", "19m"},
		{"25000", "25k"},
		{"25ms", "25ms"},
		{"cake", "CAKE"},
		{"ether1", "ether1"},
	}
	for _, p := range pairs {
		if normalizeROS(p[0]) != normalizeROS(p[1]) {
			t.Fatalf("normalizeROS(%q) != normalizeROS(%q)", p[0], p[1])
		}
	}
}

// TestRouterOSRateDisplayRounding — CHR 7.24.6 arrondit son export à une
// décimale en Mbps : un script réel demandant 16648496 bit/s s'imprime
// « 16.6Mbps ». La tolérance de 1 % doit accepter l'écart d'affichage, sans
// jamais confondre deux débits réellement différents.
func TestRouterOSRateDisplayRounding(t *testing.T) {
	ok := [][2]string{
		{"16648496", "16.6Mbps"}, // cas réel audit-1794898684 / CHR 2026-11-17
		{"19000000", "19M"},
		{"19000000", "19.0Mbps"},
	}
	for _, p := range ok {
		if !rosValuesEqual(p[0], p[1]) {
			t.Errorf("rosValuesEqual(%q,%q) = false, want true", p[0], p[1])
		}
	}
	bad := [][2]string{
		{"19000000", "20M"},     // vrai écart de configuration
		{"16648496", "cgoingress-cake"}, // chaîne non-débit
		{"25ms", "26ms"},
	}
	for _, p := range bad {
		if rosValuesEqual(p[0], p[1]) {
			t.Errorf("rosValuesEqual(%q,%q) = true, want false", p[0], p[1])
		}
	}
}

func TestRouterOSDefaultAbsentIsAccepted(t *testing.T) {
	expected := RouterOSConfig{Objects: []RouterOSObject{{
		Section: "queue type", Name: "cgo-cake",
		Props: map[string]string{"kind": "cake", "cake-flowmode": "triple-isolate"},
	}}}
	found := RouterOSConfig{Objects: []RouterOSObject{{
		Section: "queue type", Name: "cgo-cake",
		Props: map[string]string{"kind": "cake"}, // /export terse omet le défaut usine
	}}}
	checks := CompareRouterOS(expected, found)
	for _, c := range checks {
		if c.Name == "queue type cgo-cake.cake-flowmode" && !c.Pass {
			t.Fatalf("absent (défaut) refusé : %+v", c)
		}
	}
}

func TestRouterOSNonDefaultAbsentIsRefused(t *testing.T) {
	expected := RouterOSConfig{Objects: []RouterOSObject{{
		Section: "queue type", Name: "cgo-cake",
		Props: map[string]string{"cake-rtt": "25ms"},
	}}}
	found := RouterOSConfig{Objects: []RouterOSObject{{
		Section: "queue type", Name: "cgo-cake",
		Props: map[string]string{}, // 25ms n'est PAS le défaut — doit échouer
	}}}
	checks := CompareRouterOS(expected, found)
	for _, c := range checks {
		if c.Name == "queue type cgo-cake.cake-rtt" && c.Pass {
			t.Fatalf("absent non-default accepté : %+v", c)
		}
	}
}
