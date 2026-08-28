package qdisc

import "testing"

// Leaf-only goodput — receiver tc output sums parents+leaf (2-3x inflation, P1 80Mbit observed 231).
// SumBytes must return the leaf (last qdisc) counter, not the sum. Golden: 77 valid G4.
func TestStatsLeafOnly(t *testing.T) {
	stats := []Stats{{Kind: "netem", Bytes: 100}, {Kind: "tbf", Bytes: 80}, {Kind: "fq_codel", Bytes: 77}}
	if got := SumBytes(stats); got != 77 {
		t.Fatalf("leaf not 77, got %d", got)
	}
	if got := SumBytes(nil); got != 0 {
		t.Fatalf("empty not 0, got %d", got)
	}
	single := []Stats{{Kind: "tbf", Bytes: 42}}
	if got := SumBytes(single); got != 42 {
		t.Fatalf("single not 42, got %d", got)
	}
}