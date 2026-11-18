package qdisc

import "testing"

func TestSumDropsMeasuredReportsAbsence(t *testing.T) {
	if v, ok := SumDropsMeasured(nil); ok || v != 0 {
		t.Fatalf("SumDropsMeasured(nil) = (%d, %v), want (0, false)", v, ok)
	}
	if v, ok := SumDropsMeasured([]Stats{}); ok || v != 0 {
		t.Fatalf("SumDropsMeasured(empty) = (%d, %v), want (0, false)", v, ok)
	}
	v, ok := SumDropsMeasured([]Stats{{Drops: 7}})
	if !ok || v != 7 {
		t.Fatalf("SumDropsMeasured(7) = (%d, %v), want (7, true)", v, ok)
	}
}

func TestSumBytesMeasuredReportsAbsence(t *testing.T) {
	if v, ok := SumBytesMeasured(nil); ok || v != 0 {
		t.Fatalf("SumBytesMeasured(nil) = (%d, %v), want (0, false)", v, ok)
	}
	v, ok := SumBytesMeasured([]Stats{{Bytes: 11}})
	if !ok || v != 11 {
		t.Fatalf("SumBytesMeasured(11) = (%d, %v), want (11, true)", v, ok)
	}
}
