package results

import (
	"strings"
	"testing"
)

// HardwareRecommendation single source — fq_codel/cake → MikroTik, else ISP/mini-PC CAKE bridge.
func TestHardwareSingle(t *testing.T) {
	fq := HardwareRecommendation("fq_codel", "P2")
	if !strings.Contains(fq, "MikroTik") || !strings.Contains(fq, "fq_codel") {
		t.Fatalf("fq_codel not MikroTik: %q", fq)
	}
	cake := HardwareRecommendation("cake", "P1")
	if !strings.Contains(cake, "MikroTik") || !strings.Contains(cake, "CAKE") {
		t.Fatalf("cake not MikroTik/CAKE: %q", cake)
	}
	other := HardwareRecommendation("pfifo_fast", "P1")
	if !strings.Contains(other, "mini-PC") || !strings.Contains(other, "P1") {
		t.Fatalf("pfifo_fast not ISP/mini-PC: %q", other)
	}
}