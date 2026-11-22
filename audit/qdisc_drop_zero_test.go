package audit

import (
	"encoding/json"
	"strings"
	"testing"
)

// A measured zero qdisc drop count is a real observation and must survive
// serialization. With `omitempty` the key vanished and cgo kerneltruth seal
// refused the report as "absent ou null" on a genuinely zero reading.
func TestQdiscDropPacketsZeroIsSerialized(t *testing.T) {
	b, err := json.Marshal(KernelTelemetry{QdiscDropPackets: 0})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"qdisc_drop_packets":0`) {
		t.Fatalf("measured zero qdisc drops were dropped from JSON: %s", b)
	}
}

// Control: a non-zero reading is unchanged by the serialization fix.
func TestQdiscDropPacketsNonZeroRoundTrips(t *testing.T) {
	const want uint64 = 7
	b, err := json.Marshal(KernelTelemetry{QdiscDropPackets: want})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got KernelTelemetry
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.QdiscDropPackets != want {
		t.Fatalf("qdisc_drop_packets changed: got %d, want %d (JSON %s)", got.QdiscDropPackets, want, b)
	}
}
